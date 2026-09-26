// Объявление window.WebApp, которое инжектит MAX Bridge
// (https://st.max.ru/js/max-web-app.js). В обычном браузера его нет.

interface MaxUser {
  id?: number
  name?: string
  username?: string
}

interface MaxWebApp {
  initData?: string
  initDataUnsafe?: { user?: MaxUser; auth_date?: number }
  colorScheme?: string
  viewportStableHeight?: number
  ready?: () => void
  expand?: () => void
  openLink?: (url: string) => void
}

interface Window {
  WebApp?: MaxWebApp
}
