import { ref } from 'vue'

type Theme = 'light' | 'dark'

const THEME_KEY = 'theme-preference'
const currentTheme = ref<Theme>('light')

export function useTheme() {
    const setTheme = (theme: Theme) => {
        currentTheme.value = theme
        localStorage.setItem(THEME_KEY, theme)
        
        if (theme === 'dark') {
            document.documentElement.classList.add('dark')
        } else {
            document.documentElement.classList.remove('dark')
        }
    }

    const toggleTheme = () => {
        const newTheme = currentTheme.value === 'light' ? 'dark' : 'light'
        setTheme(newTheme)
    }

    const initTheme = () => {
        const savedTheme = localStorage.getItem(THEME_KEY) as Theme | null
        
        if (savedTheme) {
            setTheme(savedTheme)
        } else {
            // Automatic theme detection from the browser
            const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches
            setTheme(prefersDark ? 'dark' : 'light')
        }

        // Listening for changes in system preferences
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
            if (!localStorage.getItem(THEME_KEY)) {
                setTheme(e.matches ? 'dark' : 'light')
            }
        })
    }

    return {
        currentTheme,
        setTheme,
        toggleTheme,
        initTheme
    }
}
