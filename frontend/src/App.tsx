import { useEffect } from 'react'
import { HashRouter, Navigate, Route, Routes } from 'react-router-dom'
import { StoreProvider } from './store'
import Home from './screens/Home'
import Form from './screens/Form'
import Review from './screens/Review'
import Result from './screens/Result'
import Consequences from './screens/Consequences'
import Checklist from './screens/Checklist'
import History from './screens/History'
import MarketingForm from './screens/MarketingForm'
import MarketingResult from './screens/MarketingResult'

export default function App() {
  // MAX Bridge: сообщаем мессенджеру, что приложение отрисовано
  // (в обычном браузере WebApp нет — вызов молча пропускается).
  useEffect(() => {
    window.WebApp?.ready?.()
  }, [])

  return (
    <StoreProvider>
      <HashRouter>
        <Routes>
          <Route path="/" element={<Home />} />
          <Route path="/form" element={<Form />} />
          <Route path="/review" element={<Review />} />
          <Route path="/result" element={<Result />} />
          <Route path="/consequences" element={<Consequences />} />
          <Route path="/checklist" element={<Checklist />} />
          <Route path="/history" element={<History />} />
          <Route path="/marketing" element={<MarketingForm />} />
          <Route path="/marketing/result" element={<MarketingResult />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </HashRouter>
    </StoreProvider>
  )
}
