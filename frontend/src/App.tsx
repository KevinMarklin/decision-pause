import { HashRouter, Navigate, Route, Routes } from 'react-router-dom'
import { StoreProvider } from './store'
import Home from './screens/Home'
import Form from './screens/Form'
import Review from './screens/Review'
import Result from './screens/Result'
import Consequences from './screens/Consequences'
import Checklist from './screens/Checklist'
import History from './screens/History'

export default function App() {
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
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </HashRouter>
    </StoreProvider>
  )
}
