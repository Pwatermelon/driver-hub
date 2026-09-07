import { Navigate, Route, Routes } from 'react-router-dom'
import { HomePage } from './pages/HomePage'
import { DriverPage } from './pages/DriverPage'
import { DriverVerifyPage } from './pages/DriverVerifyPage'
import { CompanyPage } from './pages/CompanyPage'

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<HomePage />} />
      <Route path="/driver" element={<DriverPage />} />
      <Route path="/driver/verify" element={<DriverVerifyPage />} />
      <Route path="/company" element={<CompanyPage />} />
      <Route path="/driver.html" element={<Navigate to="/driver" replace />} />
      <Route path="/company.html" element={<Navigate to="/company" replace />} />
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
