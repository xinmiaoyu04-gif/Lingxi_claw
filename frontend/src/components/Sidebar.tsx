import { NavLink } from 'react-router-dom'

const navItems = [
  { to: '/home', label: 'Home' },
  { to: '/courses', label: 'Courses' },
  { to: '/analytics', label: 'Analytics' },
]

export default function Sidebar() {
  return (
    <aside className="sidebar app-shell-sidebar">
      <div className="sidebar-header">
        <div className="sidebar-logo">LINGXI</div>
        <div className="sidebar-subtitle">Study with Lingxi</div>
      </div>

      <nav className="sidebar-nav">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            className={({ isActive }) => `sidebar-item ${isActive ? 'active' : ''}`}
          >
            {item.label}
          </NavLink>
        ))}
      </nav>

      <div className="sidebar-footer">
        <NavLink
          to="/settings"
          className={({ isActive }) => `sidebar-item ${isActive ? 'active' : ''}`}
        >
          Settings
        </NavLink>
        <div className="sidebar-user">
          <div className="sidebar-user-avatar">U</div>
          <div className="sidebar-user-name">User</div>
        </div>
      </div>
    </aside>
  )
}
