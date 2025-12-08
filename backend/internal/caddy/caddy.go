package caddy

import (
	"bufio"
	"bytes"
	"caddy-proxy-manager/internal/db"
	"caddy-proxy-manager/internal/models"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	httpClient = &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 5,
			IdleConnTimeout:     30 * time.Second,
			DisableKeepAlives:   false,
		},
	}
	applyConfigMutex sync.Mutex
)

func WaitForCaddy(maxRetries int, retryInterval time.Duration) error {
	for i := 0; i < maxRetries; i++ {
		resp, err := httpClient.Get("http://localhost:2019/config/")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				return nil
			}
		}
		time.Sleep(retryInterval)
	}
	return fmt.Errorf("caddy admin API not ready after %d retries", maxRetries)
}

type CaddyConfig struct {
	Admin   map[string]interface{} `json:"admin,omitempty"`
	Logging map[string]interface{} `json:"logging,omitempty"`
	Apps    map[string]interface{} `json:"apps"`
}

func buildSelectionPolicy(policy string) map[string]interface{} {
	switch policy {
	case "round_robin":
		return map[string]interface{}{"policy": "round_robin"}
	case "least_conn":
		return map[string]interface{}{"policy": "least_conn"}
	case "ip_hash":
		return map[string]interface{}{"policy": "ip_hash"}
	case "random":
		return map[string]interface{}{"policy": "random"}
	case "random_choose":
		return map[string]interface{}{"policy": "random_choose", "choose": 2}
	case "first":
		return map[string]interface{}{"policy": "first"}
	case "uri_hash":
		return map[string]interface{}{"policy": "uri_hash"}
	case "header":
		return map[string]interface{}{"policy": "header", "field": "X-Forwarded-For"}
	case "cookie":
		return map[string]interface{}{"policy": "cookie", "name": "lb_cookie", "secret": "changeme"}
	default:
		return nil
	}
}

func GenerateConfig() (*CaddyConfig, error) {
	var hosts []models.Host
	if err := db.DB.Preload("Locations.Upstreams").Preload("Upstreams").Find(&hosts).Error; err != nil {
		return nil, err
	}
	dashboardRoutes := []map[string]interface{}{}
	apiRoute := map[string]interface{}{
		"match": []map[string]interface{}{
			{
				"path": []string{"/api/*"},
			},
		},
		"handle": []map[string]interface{}{
			{
				"handler": "reverse_proxy",
				"upstreams": []map[string]interface{}{
					{"dial": "localhost:8080"},
				},
			},
		},
	}
	dashboardRoutes = append(dashboardRoutes, apiRoute)
	staticRoute := map[string]interface{}{
		"match": []map[string]interface{}{
			{
				"file": map[string]interface{}{
					"try_files": []string{"{http.request.uri.path}"},
					"root":      "/app/public",
				},
			},
		},
		"handle": []map[string]interface{}{
			{
				"handler": "file_server",
				"root":    "/app/public",
			},
		},
	}
	dashboardRoutes = append(dashboardRoutes, staticRoute)
	spaRoute := map[string]interface{}{
		"handle": []map[string]interface{}{
			{
				"handler": "rewrite",
				"uri":     "/index.html",
			},
			{
				"handler":     "file_server",
				"root":        "/app/public",
				"index_names": []string{"index.html"},
			},
		},
	}
	dashboardRoutes = append(dashboardRoutes, spaRoute)
	publicRoutes := []map[string]interface{}{}
	for _, host := range hosts {
		if !host.IsActive {
			continue
		}

		subRoutes := []map[string]interface{}{}
		for _, loc := range host.Locations {
			var locHandler map[string]interface{}
			if len(loc.Upstreams) > 0 {
				upstreams := []map[string]interface{}{}
				for _, upstream := range loc.Upstreams {
					u := map[string]interface{}{
						"dial": upstream.Target,
					}
					if upstream.MaxFails > 0 {
						u["max_fails"] = upstream.MaxFails
					}
					if upstream.FailTimeout != "" {
						u["fail_duration"] = upstream.FailTimeout
					}
					upstreams = append(upstreams, u)
				}

				locHandler = map[string]interface{}{
					"handler":   "reverse_proxy",
					"upstreams": upstreams,
				}
				if loc.LoadBalancing != "" && len(upstreams) > 1 {
					selectionPolicy := buildSelectionPolicy(loc.LoadBalancing)
					if selectionPolicy != nil {
						locHandler["load_balancing"] = map[string]interface{}{
							"selection_policy": selectionPolicy,
						}
					}
				}
				if loc.LBTryDuration != "" {
					locHandler["lb_try_duration"] = loc.LBTryDuration
				}
				if loc.LBTryInterval != "" {
					locHandler["lb_try_interval"] = loc.LBTryInterval
				}
				if loc.HealthCheck && loc.HealthCheckPath != "" {
					healthCheck := map[string]interface{}{
						"uri": loc.HealthCheckPath,
					}
					if loc.HealthCheckInterval != "" {
						healthCheck["interval"] = loc.HealthCheckInterval
					}
					locHandler["health_checks"] = map[string]interface{}{
						"active": healthCheck,
					}
				}
			} else {
				locHandler = map[string]interface{}{
					"handler": "reverse_proxy",
					"upstreams": []map[string]interface{}{
						{"dial": loc.Target},
					},
				}
			}

			locRoute := map[string]interface{}{
				"match": []map[string]interface{}{
					{
						"path": []string{loc.Path, loc.Path + "/*"},
					},
				},
				"handle": []map[string]interface{}{locHandler},
			}
			subRoutes = append(subRoutes, locRoute)
		}
		var mainHandler map[string]interface{}

		if host.Type == "redirect" {
			code := 301
			if host.ForwardingCode != 0 {
				code = host.ForwardingCode
			}

			target := host.Target
			if !strings.HasPrefix(target, "http://") && !strings.HasPrefix(target, "https://") && !strings.HasPrefix(target, "/") {
				target = "http://" + target
			}

			mainHandler = map[string]interface{}{
				"handler": "static_response",
				"headers": map[string][]string{
					"Location": {target},
				},
				"status_code": code,
			}
		} else {
			upstreams := []map[string]interface{}{}
			if len(host.Upstreams) > 0 {
				for _, upstream := range host.Upstreams {
					u := map[string]interface{}{
						"dial": upstream.Target,
					}
					if upstream.MaxFails > 0 {
						u["max_fails"] = upstream.MaxFails
					}
					if upstream.FailTimeout != "" {
						u["fail_duration"] = upstream.FailTimeout
					}
					upstreams = append(upstreams, u)
				}
			} else if host.Target != "" {
				upstreams = append(upstreams, map[string]interface{}{
					"dial": host.Target,
				})
			}

			mainHandler = map[string]interface{}{
				"handler":   "reverse_proxy",
				"upstreams": upstreams,
			}
			if host.LoadBalancing != "" && len(upstreams) > 1 {
				selectionPolicy := buildSelectionPolicy(host.LoadBalancing)
				if selectionPolicy != nil {
					mainHandler["load_balancing"] = map[string]interface{}{
						"selection_policy": selectionPolicy,
					}
				}
			}
			if host.LBTryDuration != "" {
				mainHandler["lb_try_duration"] = host.LBTryDuration
			}
			if host.LBTryInterval != "" {
				mainHandler["lb_try_interval"] = host.LBTryInterval
			}
			if host.HealthCheck && host.HealthCheckPath != "" {
				healthCheck := map[string]interface{}{
					"uri": host.HealthCheckPath,
				}
				if host.HealthCheckInterval != "" {
					healthCheck["interval"] = host.HealthCheckInterval
				}
				mainHandler["health_checks"] = map[string]interface{}{
					"active": healthCheck,
				}
			}
		}

		mainRoute := map[string]interface{}{
			"handle": []map[string]interface{}{mainHandler},
		}
		subRoutes = append(subRoutes, mainRoute)
		chain := []map[string]interface{}{}
		if host.HSTSEnabled {
			chain = append(chain, map[string]interface{}{
				"handler": "headers",
				"response": map[string]interface{}{
					"set": map[string][]string{
						"Strict-Transport-Security": {"max-age=31536000; includeSubDomains; preload"},
					},
				},
			})
		}
		if host.CacheAssets {
			chain = append(chain, map[string]interface{}{
				"handler": "subroute",
				"routes": []map[string]interface{}{
					{
						"match": []map[string]interface{}{
							{
								"path": []string{"*.jpg", "*.jpeg", "*.png", "*.gif", "*.ico", "*.css", "*.js", "*.svg", "*.woff", "*.woff2"},
							},
						},
						"handle": []map[string]interface{}{
							{
								"handler": "headers",
								"response": map[string]interface{}{
									"set": map[string][]string{
										"Cache-Control": {"public, max-age=86400"},
									},
								},
							},
						},
					},
				},
			})
		}
		if host.BlockExploits {
			chain = append(chain, map[string]interface{}{
				"handler": "subroute",
				"routes": []map[string]interface{}{
					{
						"match": []map[string]interface{}{
							{
								"path": []string{"*.php", "*.git/*", "*.env"},
							},
						},
						"handle": []map[string]interface{}{
							{
								"handler":     "static_response",
								"status_code": 403,
							},
						},
					},
				},
			})
		}
		if host.AccessListID != nil {
			var accessList models.AccessList
			if err := db.DB.Preload("Clients").Preload("Rules").First(&accessList, *host.AccessListID).Error; err == nil {
				if len(accessList.Rules) > 0 {
					allowIPs := []string{}
					denyIPs := []string{}

					for _, rule := range accessList.Rules {
						if rule.IP == "" {
							continue
						}
						if rule.Action == "allow" {
							allowIPs = append(allowIPs, rule.IP)
						} else if rule.Action == "deny" {
							denyIPs = append(denyIPs, rule.IP)
						}
					}
					if len(denyIPs) > 0 {
						chain = append(chain, map[string]interface{}{
							"handler": "subroute",
							"routes": []map[string]interface{}{
								{
									"match": []map[string]interface{}{
										{
											"remote_ip": map[string]interface{}{
												"ranges": denyIPs,
											},
										},
									},
									"handle": []map[string]interface{}{
										{
											"handler":     "static_response",
											"status_code": 403,
											"body":        "Access Denied",
										},
									},
									"terminal": true,
								},
							},
						})
					}
					if len(allowIPs) > 0 {
						chain = append(chain, map[string]interface{}{
							"handler": "subroute",
							"routes": []map[string]interface{}{
								{
									"match": []map[string]interface{}{
										{
											"not": []map[string]interface{}{
												{
													"remote_ip": map[string]interface{}{
														"ranges": allowIPs,
													},
												},
											},
										},
									},
									"handle": []map[string]interface{}{
										{
											"handler":     "static_response",
											"status_code": 403,
											"body":        "Access Denied",
										},
									},
									"terminal": true,
								},
							},
						})
					}
				}
				accounts := map[string]string{}
				for _, client := range accessList.Clients {
					if client.Username != "" && client.Password != "" {
						accounts[client.Username] = client.Password
					}
				}

				if len(accounts) > 0 {
					accountList := []map[string]interface{}{}
					for u, p := range accounts {
						accountList = append(accountList, map[string]interface{}{
							"username": u,
							"password": p,
						})
					}

					chain = append(chain, map[string]interface{}{
						"handler": "authentication",
						"providers": map[string]interface{}{
							"http_basic": map[string]interface{}{
								"accounts": accountList,
							},
						},
					})
				}
			}
		}

		chain = append(chain, map[string]interface{}{
			"handler": "subroute",
			"routes":  subRoutes,
		})

		route := map[string]interface{}{
			"match": []map[string]interface{}{
				{
					"host": []string{host.Domain},
				},
			},
			"handle": chain,
		}
		publicRoutes = append(publicRoutes, route)
	}
	defaultPageRoute := map[string]interface{}{
		"handle": []map[string]interface{}{
			{
				"handler":     "file_server",
				"root":        "/data/default",
				"index_names": []string{"index.html"},
			},
		},
	}
	publicRoutes = append(publicRoutes, defaultPageRoute)

	httpApp := map[string]interface{}{
		"servers": map[string]interface{}{
			"srv_public": map[string]interface{}{
				"listen": []string{":443", ":80"},
				"routes": publicRoutes,
			},
			"srv_dashboard": map[string]interface{}{
				"listen": []string{":81"},
				"routes": dashboardRoutes,
			},
		},
	}
	var certs []models.Certificate
	db.DB.Find(&certs)

	tlsApp := map[string]interface{}{}
	if len(certs) > 0 {
		loadFiles := []map[string]interface{}{}
		for _, cert := range certs {
			loadFiles = append(loadFiles, map[string]interface{}{
				"certificate": cert.CertFile,
				"key":         cert.KeyFile,
				"tags":        []string{cert.Domain},
			})
		}
		tlsApp["certificates"] = map[string]interface{}{
			"load_files": loadFiles,
		}
	}
	policies := []map[string]interface{}{}
	for _, host := range hosts {
		if !host.SSL {
			continue
		}

		var issuers []map[string]interface{}
		effectiveProvider := host.SSLProvider
		if host.SSLActualProvider != "" && host.SSLActualProvider != host.SSLProvider {
			effectiveProvider = host.SSLActualProvider
		}

		switch effectiveProvider {
		case "letsencrypt":
			issuers = []map[string]interface{}{
				{"module": "acme", "email": "admin@" + host.Domain},
			}
		case "zerossl":
			if host.UseDNSChallenge && host.DNSProvider != "" {
				var kidSetting, hmacSetting models.Setting
				kidErr := db.DB.Where("key = ?", "zerossl_eab_kid").First(&kidSetting).Error
				hmacErr := db.DB.Where("key = ?", "zerossl_eab_hmac_key").First(&hmacSetting).Error

				if kidErr == nil && hmacErr == nil && kidSetting.Value != "" && hmacSetting.Value != "" {
					issuers = []map[string]interface{}{
						{
							"module": "acme",
							"ca":     "https://acme.zerossl.com/v2/DV90",
							"email":  "admin@" + host.Domain,
							"external_account": map[string]interface{}{
								"key_id":  kidSetting.Value,
								"mac_key": hmacSetting.Value,
							},
						},
					}
				} else {
					issuers = []map[string]interface{}{
						{"module": "zerossl"},
					}
				}
			} else {
				issuers = []map[string]interface{}{
					{"module": "zerossl"},
				}
			}
		case "selfsigned":
			issuers = []map[string]interface{}{
				{"module": "internal"},
			}
		case "custom":
			if host.CertificateID != nil {
				issuers = []map[string]interface{}{
					{"module": "internal"},
				}
			} else {
				continue
			}
		case "auto":
			continue
		default:
			continue
		}

		if len(issuers) > 0 {
			policy := map[string]interface{}{
				"subjects": []string{host.Domain},
				"issuers":  issuers,
			}
			if host.UseDNSChallenge && host.DNSProvider != "" {
				var providerConfig map[string]interface{}

				switch host.DNSProvider {
				case "cloudflare":
					providerConfig = map[string]interface{}{
						"name":      "cloudflare",
						"api_token": host.DNSToken,
					}
				}

				if providerConfig != nil {
					for i, issuer := range issuers {
						if issuer["module"] == "acme" {
							issuers[i]["challenges"] = map[string]interface{}{
								"dns": map[string]interface{}{
									"provider": providerConfig,
								},
							}
						}
					}
					policy["issuers"] = issuers
				}
			}

			policies = append(policies, policy)
		}
	}
	var streams []models.Stream
	if err := db.DB.Find(&streams).Error; err == nil && len(streams) > 0 {
		for _, stream := range streams {
			if stream.SSL && stream.IsActive && stream.TCPEnabled {
				if stream.CertificateID != nil {
					var cert models.Certificate
					if err := db.DB.First(&cert, *stream.CertificateID).Error; err == nil && cert.CertFile != "" {
						continue
					}
				}
				streamSubject := fmt.Sprintf("stream-%d.internal", stream.ID)
				policies = append(policies, map[string]interface{}{
					"subjects": []string{streamSubject},
					"issuers": []map[string]interface{}{
						{"module": "internal"},
					},
				})
			}
		}
	}

	if len(policies) > 0 {
		tlsApp["automation"] = map[string]interface{}{
			"policies": policies,
		}
	}

	apps := map[string]interface{}{
		"http": httpApp,
	}
	if len(tlsApp) > 0 {
		apps["tls"] = tlsApp
	}
	if len(streams) > 0 {
		layer4Servers := map[string]interface{}{}

		for _, stream := range streams {
			if !stream.IsActive {
				continue
			}
			if !stream.TCPEnabled && !stream.UDPEnabled {
				continue
			}
			proxyHandler := map[string]interface{}{
				"handler": "proxy",
				"upstreams": []map[string]interface{}{
					{"dial": []string{stream.Target}},
				},
			}
			if stream.TCPEnabled {
				tcpServerName := fmt.Sprintf("stream_%d_tcp", stream.ID)
				handlers := []map[string]interface{}{}
				if stream.SSL {
					tlsHandler := map[string]interface{}{
						"handler": "tls",
					}
					if stream.CertificateID != nil {
						var cert models.Certificate
						if err := db.DB.First(&cert, *stream.CertificateID).Error; err == nil {
							if cert.CertFile != "" && cert.KeyFile != "" {
								tlsHandler["connection_policies"] = []map[string]interface{}{
									{
										"certificate_selection": map[string]interface{}{
											"any_tag": []string{fmt.Sprintf("cert-%d", cert.ID)},
										},
									},
								}
							}
						}
					}

					handlers = append(handlers, tlsHandler)
				}
				handlers = append(handlers, proxyHandler)

				route := map[string]interface{}{
					"handle": handlers,
				}

				serverConfig := map[string]interface{}{
					"listen": []string{fmt.Sprintf("tcp/:%d", stream.ListenPort)},
					"routes": []map[string]interface{}{route},
				}

				layer4Servers[tcpServerName] = serverConfig
			}
			if stream.UDPEnabled {
				udpHandler := map[string]interface{}{
					"handler": "proxy",
					"upstreams": []map[string]interface{}{
						{"dial": []string{stream.Target}},
					},
				}
				udpRoute := map[string]interface{}{
					"handle": []map[string]interface{}{udpHandler},
				}
				udpServerName := fmt.Sprintf("stream_%d_udp", stream.ID)
				layer4Servers[udpServerName] = map[string]interface{}{
					"listen": []string{fmt.Sprintf("udp/:%d", stream.ListenPort)},
					"routes": []map[string]interface{}{udpRoute},
				}
			}
		}

		if len(layer4Servers) > 0 {
			apps["layer4"] = map[string]interface{}{
				"servers": layer4Servers,
			}
		}
	}

	logging := map[string]interface{}{
		"logs": map[string]interface{}{
			"default": map[string]interface{}{
				"writer": map[string]interface{}{
					"output":   "file",
					"filename": "/data/caddy.log",
				},
				"level": "INFO",
			},
		},
	}
	admin := map[string]interface{}{
		"listen": ":2019",
		"config": map[string]interface{}{
			"persist": false,
		},
	}

	config := &CaddyConfig{
		Admin:   admin,
		Logging: logging,
		Apps:    apps,
	}

	return config, nil
}

func ApplyConfig() error {
	applyConfigMutex.Lock()
	defer applyConfigMutex.Unlock()

	config, err := GenerateConfig()
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(config)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "http://localhost:2019/load", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to apply config: %s - %s", resp.Status, string(body))
	}

	return nil
}
func CheckSSLStatus(domain string, provider string) bool {
	baseDir := "/data/caddy/certificates"
	customDir := "/data/custom_ssl"

	checkPath := func(p string) bool {
		path := ""
		switch p {
		case "letsencrypt":
			path = fmt.Sprintf("%s/acme-v02.api.letsencrypt.org-directory/%s/%s.crt", baseDir, domain, domain)
		case "zerossl":
			path = fmt.Sprintf("%s/acme.zerossl.com-v2-dv90/%s/%s.crt", baseDir, domain, domain)
		case "selfsigned":
			path = fmt.Sprintf("%s/%s-selfsigned.crt", customDir, domain)
		}
		if path != "" {
			if _, err := os.Stat(path); err == nil {
				return true
			}
		}
		return false
	}

	if provider == "auto" {
		return checkPath("letsencrypt") || checkPath("zerossl") || checkPath("selfsigned")
	}
	return checkPath(provider)
}
func CheckSSLStatusWithProvider(domain string, provider string) (bool, string) {
	baseDir := "/data/caddy/certificates"
	customDir := "/data/custom_ssl"

	checkPath := func(p string) bool {
		path := ""
		switch p {
		case "letsencrypt":
			path = fmt.Sprintf("%s/acme-v02.api.letsencrypt.org-directory/%s/%s.crt", baseDir, domain, domain)
		case "zerossl":
			path = fmt.Sprintf("%s/acme.zerossl.com-v2-dv90/%s/%s.crt", baseDir, domain, domain)
		case "selfsigned":
			path = fmt.Sprintf("%s/%s-selfsigned.crt", customDir, domain)
		}
		if path != "" {
			if _, err := os.Stat(path); err == nil {
				return true
			}
		}
		return false
	}

	if provider == "auto" {
		if checkPath("letsencrypt") {
			return true, "letsencrypt"
		}
		if checkPath("zerossl") {
			return true, "zerossl"
		}
		if checkPath("selfsigned") {
			return true, "selfsigned"
		}
		return false, ""
	}

	if checkPath(provider) {
		return true, provider
	}
	return false, ""
}
func GetSSLError(domain string, sinceTime time.Time) (string, bool) {
	logFile := "/data/caddy.log"

	file, err := os.Open(logFile)
	if err != nil {
		return "", false
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	var lastError string
	found := false

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, `"level":"error"`) || strings.Contains(line, `"level":"warn"`) {
			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(line), &logEntry); err == nil {
				if ts, ok := logEntry["ts"].(float64); ok {
					logTime := time.Unix(int64(ts), int64((ts-float64(int64(ts)))*1e9))
					if logTime.Before(sinceTime) {
						continue
					}
				}

				identifier, _ := logEntry["identifier"].(string)
				msg, _ := logEntry["msg"].(string)
				errStr, _ := logEntry["error"].(string)
				if identifier == domain || strings.Contains(errStr, domain) || strings.Contains(msg, domain) {
					lastError = errStr
					if lastError == "" {
						lastError = msg
					}
					found = true
				}
			}
		}
		if strings.Contains(line, domain) && !strings.Contains(line, `"level":"info"`) {
			if strings.Contains(line, "Cannot issue") || strings.Contains(line, "rejectedIdentifier") || strings.Contains(line, "challenge failed") || strings.Contains(line, "acme_error") || strings.Contains(line, "context canceled") {
				var logEntry map[string]interface{}
				if err := json.Unmarshal([]byte(line), &logEntry); err == nil {
					if ts, ok := logEntry["ts"].(float64); ok {
						logTime := time.Unix(int64(ts), int64((ts-float64(int64(ts)))*1e9))
						if logTime.Before(sinceTime) {
							continue
						}
					}
				}
				if !found {
					lastError = line
					if len(lastError) > 200 {
						lastError = lastError[:200] + "..."
					}
					found = true
				}
			}
		}
	}

	return lastError, found
}
