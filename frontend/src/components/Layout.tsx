import { useState } from 'react'
import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import { Icon } from '@iconify/react'
import { useAuth } from '../contexts/AuthContext'
import { useTranslation } from '../i18n/context'
import Footer from './Footer'

export default function Layout() {
  const { user, logout } = useAuth()
  const { t } = useTranslation()
  const navigate = useNavigate()
  const location = useLocation()
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)
  const isPetPage = /^\/pets\/[^/]+(\/edit)?$/.test(location.pathname)

  async function handleLogout() {
    await logout()
    navigate('/login')
    setMobileMenuOpen(false)
  }

  function closeMobileMenu() {
    setMobileMenuOpen(false)
  }

  const navLinkClass = ({ isActive }: { isActive: boolean }) => (isActive ? 'active' : '')

  return (
    <div className="app-layout">
      <nav className="nav">
        <div className="nav-left-slot">
          {isPetPage ? (
            <NavLink to="/" className="nav-back-btn">
              <Icon icon="mdi:arrow-left" width={20} height={20} />
              <span>{t('nav.back')}</span>
            </NavLink>
          ) : (
            <NavLink to="/" className="nav-brand" onClick={closeMobileMenu}>
              <span className="nav-logo">🐾</span>
              <span className="nav-brand-text">{t('nav.brand')}</span>
            </NavLink>
          )}
        </div>
        <button
          type="button"
          className="nav-menu-btn"
          onClick={() => setMobileMenuOpen((o) => !o)}
          aria-label={mobileMenuOpen ? 'Close menu' : 'Open menu'}
          aria-expanded={mobileMenuOpen}
        >
          <Icon icon={mobileMenuOpen ? 'mdi:close' : 'mdi:menu'} width={24} height={24} />
        </button>
        <div className={`nav-links ${mobileMenuOpen ? 'nav-links-open' : ''}`}>
          <NavLink to="/" end className={navLinkClass} onClick={closeMobileMenu}>
            <Icon icon="mdi:home" width={20} height={20} />
            <span>{t('nav.home')}</span>
          </NavLink>
          <NavLink to="/pets/new" className={navLinkClass} onClick={closeMobileMenu}>
            <Icon icon="mdi:paw-plus" width={20} height={20} />
            <span>{t('nav.addPet')}</span>
          </NavLink>
          {user?.role === 'admin' && (
            <>
              <NavLink to="/users" className={navLinkClass} onClick={closeMobileMenu}>
                <Icon icon="mdi:account-group" width={20} height={20} />
                <span>{t('nav.users')}</span>
              </NavLink>
              <NavLink to="/admin/options" className={navLinkClass} onClick={closeMobileMenu}>
                <Icon icon="mdi:format-list-bulleted" width={20} height={20} />
                <span>{t('nav.defaultOptions')}</span>
              </NavLink>
            </>
          )}
          <NavLink to="/settings" className={navLinkClass} onClick={closeMobileMenu}>
            <Icon icon="mdi:cog" width={20} height={20} />
            <span>{t('nav.settings')}</span>
          </NavLink>
        </div>
        <div className={`nav-right ${mobileMenuOpen ? 'nav-right-open' : ''}`}>
          <span className="nav-user">{user?.display_name}</span>
          <span
            className={`nav-role-badge ${user?.role === 'admin' ? 'nav-role-admin' : 'nav-role-user'}`}
          >
            {user?.role === 'admin' ? t('nav.roleAdmin') : t('nav.roleUser')}
          </span>
          <button type="button" className="btn" onClick={handleLogout} title={t('nav.logout')}>
            <Icon icon="mdi:logout" width={20} height={20} />
            <span>{t('nav.logout')}</span>
          </button>
        </div>
      </nav>
      <main className="main">
        <Outlet />
      </main>
      <Footer />
    </div>
  )
}
