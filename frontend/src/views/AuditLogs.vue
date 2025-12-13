<template>
    <div class="p-4 md:p-6 max-w-6xl mx-auto">
        <!-- Header -->
        <div class="bg-white rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 mb-6 overflow-hidden">
            <div class="h-1 bg-purple-500"></div>
            <div class="p-4 md:p-6">
                <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-4">
                    <div>
                        <h1 class="text-xl font-bold text-gray-900 dark:text-gray-100 dark:text-gray-100">Audit Logs</h1>
                        <p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Track system activities and user actions</p>
                    </div>
                    <div class="flex flex-col sm:flex-row gap-2 sm:items-center">
                        <div class="relative">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                            </svg>
                            <input 
                                v-model="searchQuery" 
                                type="text" 
                                placeholder="Search logs..." 
                                class="w-full sm:w-64 pl-9 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-xl text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 placeholder-gray-400 dark:placeholder-gray-500 focus:ring-2 focus:ring-purple-500 focus:border-purple-500 transition-all outline-none"
                            />
                        </div>
                        <button @click="fetchLogs" :disabled="isLoading" class="px-4 py-2 text-sm font-medium text-white bg-purple-500 hover:bg-purple-600 disabled:opacity-70 rounded-xl shadow-lg shadow-purple-500/25 flex items-center justify-center gap-1.5 transition-all whitespace-nowrap">
                            <svg xmlns="http://www.w3.org/2000/svg" :class="['h-4 w-4 transition-transform duration-500', isLoading ? 'animate-spin' : '']" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
                            </svg>
                            {{ isLoading ? 'Loading...' : 'Refresh' }}
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <!-- Statistics Cards -->
        <div class="grid grid-cols-2 md:grid-cols-5 gap-3 mb-6">
            <button 
                @click="activeFilter = null" 
                :class="['relative bg-white dark:bg-gray-800 rounded-xl border p-3 transition-all duration-300 hover:shadow-md group', activeFilter === null ? 'border-purple-500 ring-2 ring-purple-500/20' : 'border-gray-200 dark:border-gray-700']"
            >
                <div class="flex items-center gap-2">
                    <div class="w-8 h-8 rounded-lg bg-purple-500 flex items-center justify-center shadow-lg shadow-purple-500/25">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                        </svg>
                    </div>
                    <div class="text-left">
                        <p class="text-lg font-bold text-gray-900 dark:text-gray-100 dark:text-gray-100">{{ stats.total }}</p>
                        <p class="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-400 dark:text-gray-500">All Logs</p>
                    </div>
                </div>
            </button>
            <button 
                @click="activeFilter = 'create'" 
                :class="['relative bg-white dark:bg-gray-800 rounded-xl border p-3 transition-all duration-300 hover:shadow-md group', activeFilter === 'create' ? 'border-green-500 ring-2 ring-green-500/20' : 'border-gray-200 dark:border-gray-700']"
            >
                <div class="flex items-center gap-2">
                    <div class="w-8 h-8 rounded-lg bg-green-500 flex items-center justify-center shadow-lg shadow-green-500/25">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                        </svg>
                    </div>
                    <div class="text-left">
                        <p class="text-lg font-bold text-gray-900 dark:text-gray-100 dark:text-gray-100">{{ stats.created }}</p>
                        <p class="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-400 dark:text-gray-500">Created</p>
                    </div>
                </div>
            </button>
            <button 
                @click="activeFilter = 'update'" 
                :class="['relative bg-white dark:bg-gray-800 rounded-xl border p-3 transition-all duration-300 hover:shadow-md group', activeFilter === 'update' ? 'border-yellow-500 ring-2 ring-yellow-500/20' : 'border-gray-200 dark:border-gray-700']"
            >
                <div class="flex items-center gap-2">
                    <div class="w-8 h-8 rounded-lg bg-yellow-500 flex items-center justify-center shadow-lg shadow-yellow-500/25">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                    </div>
                    <div class="text-left">
                        <p class="text-lg font-bold text-gray-900 dark:text-gray-100 dark:text-gray-100">{{ stats.updated }}</p>
                        <p class="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-400 dark:text-gray-500">Updated</p>
                    </div>
                </div>
            </button>
            <button 
                @click="activeFilter = 'delete'" 
                :class="['relative bg-white dark:bg-gray-800 rounded-xl border p-3 transition-all duration-300 hover:shadow-md group', activeFilter === 'delete' ? 'border-red-500 ring-2 ring-red-500/20' : 'border-gray-200 dark:border-gray-700']"
            >
                <div class="flex items-center gap-2">
                    <div class="w-8 h-8 rounded-lg bg-red-500 flex items-center justify-center shadow-lg shadow-red-500/25">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                    </div>
                    <div class="text-left">
                        <p class="text-lg font-bold text-gray-900 dark:text-gray-100 dark:text-gray-100">{{ stats.deleted }}</p>
                        <p class="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-400 dark:text-gray-500">Deleted</p>
                    </div>
                </div>
            </button>
            <button 
                @click="activeFilter = 'login'" 
                :class="['relative bg-white dark:bg-gray-800 rounded-xl border p-3 transition-all duration-300 hover:shadow-md group', activeFilter === 'login' ? 'border-blue-500 ring-2 ring-blue-500/20' : 'border-gray-200 dark:border-gray-700']"
            >
                <div class="flex items-center gap-2">
                    <div class="w-8 h-8 rounded-lg bg-blue-500 flex items-center justify-center shadow-lg shadow-blue-500/25">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
                        </svg>
                    </div>
                    <div class="text-left">
                        <p class="text-lg font-bold text-gray-900 dark:text-gray-100 dark:text-gray-100">{{ stats.logins }}</p>
                        <p class="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-400 dark:text-gray-500">Logins</p>
                    </div>
                </div>
            </button>
        </div>

        <!-- List -->
        <div class="bg-white rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
            <!-- Loading Skeleton -->
            <div v-if="isLoading" class="divide-y divide-gray-100 dark:divide-gray-700 bg-white dark:bg-gray-800">
                <div v-for="i in 5" :key="i" class="p-4 animate-pulse">
                    <div class="flex items-center gap-4">
                        <div class="w-10 h-10 rounded-xl bg-gray-200 dark:bg-gray-700"></div>
                        <div class="flex-1">
                            <div class="h-4 bg-gray-200 dark:bg-gray-700 rounded w-1/3 mb-2"></div>
                            <div class="h-3 bg-gray-100 dark:bg-gray-700/50 rounded w-2/3"></div>
                        </div>
                        <div class="h-8 w-24 bg-gray-200 dark:bg-gray-700 rounded-lg"></div>
                    </div>
                </div>
            </div>

            <!-- Empty State: No Logs -->
            <div v-else-if="logs.length === 0 && !isLoading" class="p-12 text-center bg-white dark:bg-gray-800">
                <div class="mb-4">
                    <div class="mx-auto w-20 h-20 rounded-2xl bg-purple-500 flex items-center justify-center shadow-lg shadow-purple-500/25">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                        </svg>
                    </div>
                </div>
                <p class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-1">No audit logs yet</p>
                <p class="text-sm text-gray-500 dark:text-gray-400 dark:text-gray-400 dark:text-gray-500">Activity logs will appear here as actions are performed</p>
            </div>
            
            <!-- Empty State: No Search Results -->
            <div v-else-if="filteredLogs.length === 0 && !isLoading" class="p-12 text-center bg-white dark:bg-gray-800">
                <div class="mb-4">
                    <div class="mx-auto w-20 h-20 rounded-2xl bg-gray-100 dark:bg-gray-700 flex items-center justify-center">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                        </svg>
                    </div>
                </div>
                <p class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-1">No results found</p>
                <p class="text-sm text-gray-500 dark:text-gray-400 dark:text-gray-400 dark:text-gray-500">No logs match "<span class="font-medium text-purple-600">{{ searchQuery }}</span>"</p>
            </div>

            <!-- Log Items -->
            <template v-else>
                <!-- Desktop View -->
                <div class="hidden lg:block divide-y divide-gray-100 dark:divide-gray-700">
                    <TransitionGroup name="log-list">
                    <div v-for="(log, index) in filteredLogs" :key="log.id" 
                        class="group p-4 bg-white dark:bg-gray-800 hover:bg-gray-50/50 transition-all duration-200"
                        :style="{ animationDelay: `${index * 50}ms` }">
                        <div class="flex items-center gap-4">
                            <!-- Icon with gradient background based on action -->
                            <div class="flex-shrink-0">
                                <div :class="[
                                    'w-10 h-10 rounded-xl flex items-center justify-center shadow-lg transition-all group-hover:scale-110',
                                    log.action.includes('create') ? 'bg-green-500 shadow-green-500/25' :
                                    log.action.includes('delete') ? 'bg-red-500 shadow-red-500/25' :
                                    log.action.includes('update') || log.action.includes('change') ? 'bg-yellow-500 shadow-yellow-500/25' :
                                    log.action.includes('login') ? 'bg-blue-500 shadow-blue-500/25' :
                                    'bg-gray-400 shadow-gray-500/25'
                                ]">
                                    <svg v-if="log.action.includes('create')" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                                    </svg>
                                    <svg v-else-if="log.action.includes('delete')" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                                    </svg>
                                    <svg v-else-if="log.action.includes('update') || log.action.includes('change')" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                                    </svg>
                                    <svg v-else-if="log.action.includes('login')" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
                                    </svg>
                                    <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                                    </svg>
                                </div>
                            </div>

                            <!-- Content -->
                            <div class="flex-1 min-w-0">
                                <div class="flex items-center gap-2 mb-0.5">
                                    <p class="text-sm font-semibold text-gray-900 dark:text-gray-100 dark:text-gray-100">{{ formatAction(log.action) }}</p>
                                    <span class="text-gray-300">—</span>
                                    <p class="text-sm text-gray-600 dark:text-gray-300 truncate">{{ getLogSummary(log.details) }}</p>
                                </div>
                                <div class="flex items-center gap-3 text-xs text-gray-500 dark:text-gray-400 dark:text-gray-400 dark:text-gray-500">
                                    <span class="flex items-center gap-1">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                                        </svg>
                                        {{ formatDate(log.created_at) }}
                                    </span>
                                    <span class="flex items-center gap-1">
                                        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
                                        </svg>
                                        {{ log.username }}
                                    </span>
                                    <span class="flex items-center gap-1 font-mono text-xs bg-gray-100 px-1.5 py-0.5 rounded">
                                        {{ log.ip_address }}
                                    </span>
                                </div>
                            </div>

                            <!-- Action Button -->
                            <div class="flex-shrink-0">
                                <button @click="openDetails(log)" class="px-3 py-1.5 text-xs font-medium text-purple-600 dark:text-purple-400 hover:text-white hover:bg-purple-500 border border-purple-200 dark:border-purple-700 hover:border-transparent rounded-lg transition-all duration-200">
                                    View Details
                                </button>
                            </div>
                        </div>
                    </div>
                    </TransitionGroup>
                </div>

                <!-- Mobile View -->
                <div class="lg:hidden divide-y divide-gray-100 dark:divide-gray-700 bg-white dark:bg-gray-800">
                    <TransitionGroup name="log-list">
                    <div v-for="(log, index) in filteredLogs" :key="log.id" class="p-4" :style="{ animationDelay: `${index * 50}ms` }">
                        <div class="flex items-start gap-3">
                            <!-- Status Strip Icon -->
                            <div :class="[
                                'flex-shrink-0 w-1.5 h-full min-h-[60px] rounded-full',
                                log.action.includes('create') ? 'bg-green-500' :
                                log.action.includes('delete') ? 'bg-red-500' :
                                log.action.includes('update') || log.action.includes('change') ? 'bg-yellow-500' :
                                log.action.includes('login') ? 'bg-blue-500' :
                                'bg-gray-400'
                            ]"></div>

                            <!-- Content -->
                            <div class="flex-1 min-w-0">
                                <div class="flex items-center gap-2 mb-1">
                                    <span :class="[
                                        'inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium',
                                        log.action.includes('create') ? 'bg-green-100 text-green-700' :
                                        log.action.includes('delete') ? 'bg-red-100 text-red-700' :
                                        log.action.includes('update') || log.action.includes('change') ? 'bg-yellow-100 text-yellow-700' :
                                        log.action.includes('login') ? 'bg-blue-100 text-blue-700' :
                                        'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300'
                                    ]">
                                        {{ formatAction(log.action) }}
                                    </span>
                                </div>
                                <p class="text-sm text-gray-700 dark:text-gray-300 mb-2 line-clamp-2">{{ getLogSummary(log.details) }}</p>
                                <div class="flex flex-wrap items-center gap-2 text-xs text-gray-500 dark:text-gray-400 dark:text-gray-400 dark:text-gray-500">
                                    <span>{{ formatDate(log.created_at) }}</span>
                                    <span class="w-1 h-1 rounded-full bg-gray-300"></span>
                                    <span>{{ log.username }}</span>
                                </div>
                                <div class="mt-3">
                                    <button @click="openDetails(log)" class="w-full px-3 py-2 text-xs font-medium text-white bg-purple-500 hover:bg-purple-600 rounded-lg shadow-lg shadow-purple-500/25 transition-all">
                                        View Details
                                    </button>
                                </div>
                            </div>
                        </div>
                    </div>
                    </TransitionGroup>
                </div>
            </template>
        </div>

        <!-- Details Modal -->
        <Transition name="modal">
        <div v-if="selectedLog" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4" @click.self="selectedLog = null">
            <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-lg overflow-hidden border border-gray-200 dark:border-gray-700 transform transition-all animate-modal-in">
                <!-- Modal Header with Action-based Color -->
                <div :class="[
                    'px-6 py-5 flex justify-between items-center relative overflow-hidden',
                    selectedLog.action.includes('create') ? 'bg-green-500' :
                    selectedLog.action.includes('delete') ? 'bg-red-500' :
                    selectedLog.action.includes('update') || selectedLog.action.includes('change') ? 'bg-yellow-500' :
                    selectedLog.action.includes('login') ? 'bg-blue-500' :
                    'bg-purple-500'
                ]">
                    <!-- Background Pattern -->
                    <div class="absolute inset-0 opacity-10">
                        <svg class="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="none">
                            <defs>
                                <pattern id="grid" width="10" height="10" patternUnits="userSpaceOnUse">
                                    <path d="M 10 0 L 0 0 0 10" fill="none" stroke="white" stroke-width="0.5"/>
                                </pattern>
                            </defs>
                            <rect width="100" height="100" fill="url(#grid)"/>
                        </svg>
                    </div>
                    
                    <div class="flex items-center gap-4 relative">
                        <!-- Action Icon -->
                        <div class="w-12 h-12 bg-white/20 backdrop-blur rounded-xl flex items-center justify-center ring-2 ring-white/30">
                            <svg v-if="selectedLog.action.includes('create')" xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
                            </svg>
                            <svg v-else-if="selectedLog.action.includes('delete')" xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                            </svg>
                            <svg v-else-if="selectedLog.action.includes('update') || selectedLog.action.includes('change')" xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                            </svg>
                            <svg v-else-if="selectedLog.action.includes('login')" xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
                            </svg>
                            <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                            </svg>
                        </div>
                        <div>
                            <h3 class="text-lg font-bold text-white">{{ formatAction(selectedLog.action) }}</h3>
                            <p class="text-white/80 text-sm">{{ getRelativeTime(selectedLog.created_at) }}</p>
                        </div>
                    </div>
                    <button @click="selectedLog = null" class="relative text-white/80 hover:text-white hover:bg-white/20 rounded-xl p-2 transition-all">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>
                
                <!-- Modal Body -->
                <div class="p-6 space-y-5 max-h-[60vh] overflow-y-auto">
                    <!-- Quick Info Cards -->
                    <div class="grid grid-cols-2 gap-3">
                        <!-- User Card -->
                        <div class="bg-gradient-to-br from-purple-50 to-purple-100/50 dark:from-purple-900/20 dark:to-purple-800/10 rounded-xl p-4 border border-purple-100 dark:border-purple-800">
                            <div class="flex items-center gap-3">
                                <div :class="[
                                    'w-10 h-10 rounded-xl flex items-center justify-center text-white font-bold shadow-lg',
                                    selectedLog.action.includes('create') ? 'bg-green-500 shadow-green-500/25' :
                                    selectedLog.action.includes('delete') ? 'bg-red-500 shadow-red-500/25' :
                                    selectedLog.action.includes('update') || selectedLog.action.includes('change') ? 'bg-yellow-500 shadow-yellow-500/25' :
                                    selectedLog.action.includes('login') ? 'bg-blue-500 shadow-blue-500/25' :
                                    'bg-purple-500 shadow-purple-500/25'
                                ]">
                                    {{ selectedLog.username.charAt(0).toUpperCase() }}
                                </div>
                                <div class="min-w-0">
                                    <p class="text-xs text-purple-600 dark:text-purple-400 font-medium">User</p>
                                    <p class="text-gray-900 dark:text-gray-100 font-semibold truncate">{{ selectedLog.username }}</p>
                                </div>
                            </div>
                        </div>
                        
                        <!-- IP Card -->
                        <div class="bg-gradient-to-br from-gray-50 to-gray-100/50 dark:from-gray-800 dark:to-gray-700/50 rounded-xl p-4 border border-gray-200 dark:border-gray-700">
                            <div class="flex items-center gap-3">
                                <div class="w-10 h-10 rounded-xl bg-gray-200 dark:bg-gray-700 flex items-center justify-center">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-gray-600 dark:text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                                    </svg>
                                </div>
                                <div class="min-w-0">
                                    <p class="text-xs text-gray-500 dark:text-gray-400 font-medium">IP Address</p>
                                    <p class="text-gray-900 dark:text-gray-100 font-mono text-sm truncate">{{ selectedLog.ip_address }}</p>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Timeline Style Date -->
                    <div class="relative">
                        <div class="flex items-center gap-3 bg-gray-50 dark:bg-gray-700 rounded-xl p-4 border border-gray-100 dark:border-gray-700">
                            <div class="w-10 h-10 rounded-xl bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center flex-shrink-0">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-purple-600 dark:text-purple-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                                </svg>
                            </div>
                            <div class="flex-1 min-w-0">
                                <p class="text-gray-900 dark:text-gray-100 font-medium">{{ formatDateFull(selectedLog.created_at) }}</p>
                                <p class="text-purple-600 dark:text-purple-400 text-sm">{{ getRelativeTime(selectedLog.created_at) }}</p>
                            </div>
                            <button @click="copyToClipboard(formatDateFull(selectedLog.created_at))" class="p-2 text-gray-400 dark:text-gray-500 hover:text-purple-600 dark:hover:text-purple-400 hover:bg-purple-50 dark:hover:bg-purple-900/20 rounded-lg transition-all flex-shrink-0" title="Copy">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                                </svg>
                            </button>
                        </div>
                    </div>

                    <!-- User Agent with Better Styling -->
                    <div v-if="selectedLog.user_agent">
                        <div class="flex items-center justify-between mb-2">
                            <span class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Device Information</span>
                        </div>
                        <div class="bg-gray-50 dark:bg-gray-700 rounded-xl border border-gray-100 dark:border-gray-700 overflow-hidden">
                            <div class="flex items-center gap-2 p-3 border-b border-gray-100 dark:border-gray-700">
                                <span v-if="parsedUserAgent.browser" class="inline-flex items-center gap-1.5 bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400 px-3 py-1.5 rounded-lg text-xs font-medium">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9" />
                                    </svg>
                                    {{ parsedUserAgent.browser }}
                                </span>
                                <span v-if="parsedUserAgent.os" class="inline-flex items-center gap-1.5 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 px-3 py-1.5 rounded-lg text-xs font-medium">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.75 17L9 20l-1 1h8l-1-1-.75-3M3 13h18M5 17h14a2 2 0 002-2V5a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
                                    </svg>
                                    {{ parsedUserAgent.os }}
                                </span>
                                <span v-if="parsedUserAgent.device" class="inline-flex items-center gap-1.5 bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-400 px-3 py-1.5 rounded-lg text-xs font-medium">
                                    <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
                                    </svg>
                                    {{ parsedUserAgent.device }}
                                </span>
                            </div>
                            <div class="p-3">
                                <p class="text-gray-500 dark:text-gray-400 text-xs font-mono leading-relaxed" style="word-break: break-word;">{{ selectedLog.user_agent }}</p>
                            </div>
                        </div>
                    </div>

                    <!-- Details Section -->
                    <div>
                        <div class="flex items-center justify-between mb-2">
                            <span class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Activity Details</span>
                            <button @click="copyToClipboard(selectedLog.details)" class="text-xs text-purple-600 dark:text-purple-400 hover:text-purple-700 dark:hover:text-purple-300 font-medium flex items-center gap-1">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                                </svg>
                                Copy
                            </button>
                        </div>
                        <div class="bg-gray-900 dark:bg-gray-950 rounded-xl p-4 text-sm text-gray-100 dark:text-gray-300 font-mono whitespace-pre-wrap overflow-x-auto border border-gray-700 dark:border-gray-800">{{ selectedLog.details }}</div>
                    </div>

                    <!-- Parsed Information with Better Styling -->
                    <div v-if="parsedDetails && Object.keys(parsedDetails).length > 0">
                        <span class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">Extracted Data</span>
                        <div class="mt-2 grid gap-2">
                            <div v-for="(value, key) in parsedDetails" :key="key" 
                                class="flex items-center justify-between bg-gray-50 dark:bg-gray-700 rounded-lg px-4 py-3 border border-gray-100 dark:border-gray-700 hover:border-purple-200 dark:hover:border-purple-700 hover:bg-purple-50/50 dark:hover:bg-purple-900/20 transition-all group">
                                <span class="text-sm text-gray-600 dark:text-gray-300 group-hover:text-purple-600 dark:group-hover:text-purple-400 transition-colors">{{ formatKey(key as string) }}</span>
                                <span class="text-sm font-semibold text-gray-900 dark:text-gray-100 bg-white dark:bg-gray-800 px-3 py-1 rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm">{{ value }}</span>
                            </div>
                        </div>
                    </div>
                </div>
                
                <!-- Modal Footer -->
                <div class="bg-gray-50 dark:bg-gray-700 px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex items-center justify-between">
                    <div class="text-xs text-gray-500 dark:text-gray-400 dark:text-gray-400 dark:text-gray-500">
                        Log ID: <span class="font-mono font-medium">{{ selectedLog.id }}</span>
                    </div>
                    <button @click="selectedLog = null" class="px-5 py-2 text-sm font-medium text-white bg-purple-500 hover:bg-purple-600 rounded-xl transition-all shadow-lg shadow-purple-500/25">
                        Close
                    </button>
                </div>
            </div>
        </div>
        </Transition>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useWebSocket } from '../composables/useWebSocket'
import { toast } from 'vue-sonner'
import { authFetch } from '../composables/useAuth'

const { on, off } = useWebSocket()

interface AuditLog {
    id: number
    user_id: number
    username: string
    action: string
    details: string
    ip_address: string
    user_agent?: string
    created_at: string
}

const logs = ref<AuditLog[]>([])
const selectedLog = ref<AuditLog | null>(null)
const searchQuery = ref('')
const isLoading = ref(false)
const activeFilter = ref<string | null>(null)

const stats = computed(() => {
    return {
        total: logs.value.length,
        created: logs.value.filter(l => l.action.includes('create')).length,
        updated: logs.value.filter(l => l.action.includes('update') || l.action.includes('change')).length,
        deleted: logs.value.filter(l => l.action.includes('delete')).length,
        logins: logs.value.filter(l => l.action.includes('login')).length
    }
})

const filteredLogs = computed(() => {
    let result = logs.value
    
    if (activeFilter.value) {
        if (activeFilter.value === 'update') {
            result = result.filter(log => log.action.includes('update') || log.action.includes('change'))
        } else {
            result = result.filter(log => log.action.includes(activeFilter.value!))
        }
    }
    
    if (searchQuery.value.trim()) {
        const q = searchQuery.value.toLowerCase()
        result = result.filter(log => 
            log.action?.toLowerCase().includes(q) ||
            log.username?.toLowerCase().includes(q) ||
            log.ip_address?.toLowerCase().includes(q) ||
            log.details?.toLowerCase().includes(q)
        )
    }
    
    return result
})

const fetchLogs = async () => {
    isLoading.value = true
    try {
        const res = await authFetch('/api/audit-logs')
        if (res.ok) {
            logs.value = await res.json()
        }
    } finally {
        setTimeout(() => {
            isLoading.value = false
        }, 300)
    }
}

const formatAction = (action: string) => {
    return action.split('_').map(word => word.charAt(0).toUpperCase() + word.slice(1)).join(' ')
}

const getLogSummary = (details: string) => {
    if (!details) return ''
    if (details.includes(': ')) {
        const parts = details.split(': ')
        if (parts.length > 1) {
            let summary = parts.slice(1).join(': ')
            if (summary.includes(' (')) {
                summary = summary.split(' (')[0]
            }
            return summary.trim()
        }
    }
    return details
}

const formatDate = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('en-US', {
        weekday: 'long',
        year: 'numeric',
        month: 'numeric',
        day: 'numeric',
        hour: 'numeric',
        minute: 'numeric',
        second: 'numeric'
    })
}

const formatDateFull = (dateStr: string) => {
    return new Date(dateStr).toLocaleString('en-US', {
        weekday: 'long',
        year: 'numeric',
        month: 'long',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    })
}

const getRelativeTime = (dateStr: string) => {
    const date = new Date(dateStr)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffSecs = Math.floor(diffMs / 1000)
    const diffMins = Math.floor(diffSecs / 60)
    const diffHours = Math.floor(diffMins / 60)
    const diffDays = Math.floor(diffHours / 24)

    if (diffSecs < 60) return 'Just now'
    if (diffMins < 60) return `${diffMins} minute${diffMins > 1 ? 's' : ''} ago`
    if (diffHours < 24) return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`
    if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`
    return formatDate(dateStr)
}

const copyToClipboard = async (text: string) => {
    try {
        await navigator.clipboard.writeText(text)
        toast.success('Copied to clipboard!')
    } catch (e) {
        toast.error('Failed to copy')
    }
}

const parseUserAgent = (ua: string) => {
    const result: { browser: string | null; os: string | null; device: string | null } = {
        browser: null,
        os: null,
        device: null
    }

    if (!ua) return result

    if (ua.includes('Chrome') && !ua.includes('Edg')) {
        const match = ua.match(/Chrome\/([\d.]+)/)
        result.browser = match ? `Chrome ${match[1].split('.')[0]}` : 'Chrome'
    } else if (ua.includes('Firefox')) {
        const match = ua.match(/Firefox\/([\d.]+)/)
        result.browser = match ? `Firefox ${match[1].split('.')[0]}` : 'Firefox'
    } else if (ua.includes('Safari') && !ua.includes('Chrome')) {
        const match = ua.match(/Version\/([\d.]+)/)
        result.browser = match ? `Safari ${match[1].split('.')[0]}` : 'Safari'
    } else if (ua.includes('Edg')) {
        const match = ua.match(/Edg\/([\d.]+)/)
        result.browser = match ? `Edge ${match[1].split('.')[0]}` : 'Edge'
    } else if (ua.includes('Opera') || ua.includes('OPR')) {
        result.browser = 'Opera'
    }

    if (ua.includes('Windows NT 10')) result.os = 'Windows 10/11'
    else if (ua.includes('Windows NT 6.3')) result.os = 'Windows 8.1'
    else if (ua.includes('Windows NT 6.2')) result.os = 'Windows 8'
    else if (ua.includes('Windows NT 6.1')) result.os = 'Windows 7'
    else if (ua.includes('Mac OS X')) {
        const match = ua.match(/Mac OS X ([\d_]+)/)
        result.os = match ? `macOS ${match[1].replace(/_/g, '.')}` : 'macOS'
    }
    else if (ua.includes('Linux')) result.os = 'Linux'
    else if (ua.includes('Android')) {
        const match = ua.match(/Android ([\d.]+)/)
        result.os = match ? `Android ${match[1]}` : 'Android'
    }
    else if (ua.includes('iOS') || ua.includes('iPhone') || ua.includes('iPad')) {
        result.os = 'iOS'
    }

    if (ua.includes('Mobile')) result.device = 'Mobile'
    else if (ua.includes('Tablet') || ua.includes('iPad')) result.device = 'Tablet'
    else result.device = 'Desktop'

    return result
}

const parsedUserAgent = computed(() => {
    if (!selectedLog.value?.user_agent) return { browser: null, os: null, device: null }
    return parseUserAgent(selectedLog.value.user_agent)
})

const parsedDetails = computed(() => {
    if (!selectedLog.value?.details) return {}
    
    const details = selectedLog.value.details
    const result: Record<string, string> = {}
    
    const hostMatch = details.match(/host[s]?:\s*([^\s(]+)/)
    if (hostMatch) result['Domain'] = hostMatch[1]
    
    const targetMatch = details.match(/Target:\s*([^,)]+)/)
    if (targetMatch) result['Target'] = targetMatch[1].trim()
    
    const typeMatch = details.match(/Type:\s*([^,)]+)/)
    if (typeMatch) result['Type'] = typeMatch[1].trim()
    
    const portMatch = details.match(/Port:\s*(\d+)/)
    if (portMatch) result['Port'] = portMatch[1]
    
    const protocolMatch = details.match(/\/(TCP|UDP)/)
    if (protocolMatch) result['Protocol'] = protocolMatch[1]
    
    const userMatch = details.match(/user:\s*([^\s(]+)/i)
    if (userMatch) result['Username'] = userMatch[1]
    
    const roleMatch = details.match(/role:\s*([^\s(]+)/i)
    if (roleMatch) result['Role'] = roleMatch[1]
    
    const certMatch = details.match(/certificate for:\s*([^\s(]+)/i)
    if (certMatch) result['Certificate Domain'] = certMatch[1]
    
    const accessListMatch = details.match(/access list:\s*([^(]+)/i)
    if (accessListMatch) result['Access List'] = accessListMatch[1].trim()
    
    const usersCountMatch = details.match(/(\d+)\s*users?/i)
    if (usersCountMatch) result['Users Count'] = usersCountMatch[1]
    
    const rulesCountMatch = details.match(/(\d+)\s*rules?/i)
    if (rulesCountMatch) result['Rules Count'] = rulesCountMatch[1]
    
    return result
})

const formatKey = (key: string) => {
    return key.replace(/_/g, ' ')
}

const openDetails = (log: AuditLog) => {
    selectedLog.value = log
}

const handleAuditLogCreated = (newLog: AuditLog) => {
    logs.value.unshift(newLog)
}

onMounted(() => {
    fetchLogs()
    on('audit_log_created', handleAuditLogCreated)
})

onUnmounted(() => {
    off('audit_log_created', handleAuditLogCreated)
})
</script>

<style scoped>
/* Log list transition animations - ultra fast & smooth */
.log-list-enter-active {
    animation: slideIn 0.3s ease-out forwards;
}

.log-list-leave-active {
    animation: slideOut 0.3s ease-in-out forwards;
    position: absolute;
    width: 100%;
    z-index: 1; /* Ensure leaving items stay on top during cross-fade */
}

.log-list-move {
    transition: transform 0.3s ease-in-out;
}

@keyframes slideIn {
    from {
        opacity: 0;
        transform: translateY(6px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}

@keyframes slideOut {
    from {
        opacity: 1;
    }
    to {
        opacity: 0;
    }
}

/* Staggered animation for items - minimal delays */
.log-list-enter-active:nth-child(1) { animation-delay: 0ms; }
.log-list-enter-active:nth-child(2) { animation-delay: 15ms; }
.log-list-enter-active:nth-child(3) { animation-delay: 30ms; }
.log-list-enter-active:nth-child(4) { animation-delay: 45ms; }
.log-list-enter-active:nth-child(5) { animation-delay: 60ms; }
.log-list-enter-active:nth-child(6) { animation-delay: 75ms; }
.log-list-enter-active:nth-child(7) { animation-delay: 90ms; }
.log-list-enter-active:nth-child(8) { animation-delay: 105ms; }
.log-list-enter-active:nth-child(9) { animation-delay: 120ms; }
.log-list-enter-active:nth-child(10) { animation-delay: 135ms; }

/* Ensure leave animations are not staggered */
.log-list-leave-active { animation-delay: 0ms !important; }

/* Modal transitions - snappy */
.modal-enter-active,
.modal-leave-active {
    transition: opacity 0.15s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-enter-from,
.modal-leave-to {
    opacity: 0;
}

.modal-enter-active .animate-modal-in,
.modal-leave-active .animate-modal-in {
    transition: transform 0.2s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.15s cubic-bezier(0.16, 1, 0.3, 1);
}

.modal-enter-from .animate-modal-in {
    opacity: 0;
    transform: scale(0.97) translateY(8px);
}

.modal-leave-to .animate-modal-in {
    opacity: 0;
    transform: scale(0.98) translateY(4px);
}

/* Modal body animation */
@keyframes modalIn {
    from {
        opacity: 0;
        transform: scale(0.97) translateY(8px);
    }
    to {
        opacity: 1;
        transform: scale(1) translateY(0);
    }
}

.animate-modal-in {
    animation: modalIn 0.2s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

/* Ultra smooth hover effects */
.log-item {
    transition: all 0.12s cubic-bezier(0.16, 1, 0.3, 1);
}

.log-item:hover {
    transform: translateX(2px);
}

/* Stats cards animation */
.grid > button {
    transition: all 0.12s cubic-bezier(0.16, 1, 0.3, 1);
}

/* Page load animation */
.max-w-6xl {
    animation: pageIn 0.2s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}

@keyframes pageIn {
    from {
        opacity: 0;
        transform: translateY(8px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}
</style>
