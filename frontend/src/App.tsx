import { lazy, Suspense } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import { AuthProvider, useAuth } from './contexts/AuthContext'
import { PWAInstallProvider } from './contexts/PWAInstallContext'
import { I18nProvider, useTranslation } from './i18n/context'
import InstallPrompt from './components/InstallPrompt'
import Layout from './components/Layout'

const Login = lazy(() => import('./pages/Login'))
const Dashboard = lazy(() => import('./pages/Dashboard'))
const PetDetail = lazy(() => import('./pages/PetDetail'))
const PetForm = lazy(() => import('./pages/PetForm'))
const Users = lazy(() => import('./pages/Users'))
const AdminDefaultOptions = lazy(() => import('./pages/AdminDefaultOptions'))
const Settings = lazy(() => import('./pages/Settings'))

function LoginOrRedirect() {
  const { user, loading } = useAuth()
  const { t } = useTranslation()
  if (loading) return <div className="loading-screen">{t('common.loading')}</div>
  if (user) return <Navigate to="/" replace />
  return <Login />
}

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { user, loading } = useAuth()
  const { t } = useTranslation()
  if (loading) return <div className="loading-screen">{t('common.loading')}</div>
  if (!user) return <Navigate to="/login" replace />
  return <>{children}</>
}

function PageFallback() {
  const { t } = useTranslation()
  return <div className="loading-screen">{t('common.loading')}</div>
}

function AppRoutes() {
  return (
    <Suspense fallback={<PageFallback />}>
      <Routes>
        <Route path="/login" element={<LoginOrRedirect />} />
        <Route
          path="/"
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route index element={<Dashboard />} />
          <Route path="pets/new" element={<PetForm />} />
          <Route path="pets/:id" element={<PetDetail />} />
          <Route path="pets/:id/edit" element={<PetForm />} />
          <Route path="users" element={<Users />} />
          <Route path="admin/options" element={<AdminDefaultOptions />} />
          <Route path="settings" element={<Settings />} />
          <Route path="settings/:userId" element={<Settings />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </Suspense>
  )
}

function AppWithI18n() {
  const { user } = useAuth()
  return (
    <I18nProvider defaultLanguage={user?.language ?? 'en'}>
      <AppRoutes />
    </I18nProvider>
  )
}

export default function App() {
  return (
    <BrowserRouter>
      <PWAInstallProvider>
        <InstallPrompt />
        <AuthProvider>
          <AppWithI18n />
        </AuthProvider>
      </PWAInstallProvider>
    </BrowserRouter>
  )
}
