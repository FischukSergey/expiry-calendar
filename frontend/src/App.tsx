import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'

import { AdminRoute, GuestRoute, ProtectedRoute } from './components/ProtectedRoute.tsx'
import { Layout } from './components/Layout.tsx'
import { AuditPage } from './pages/AuditPage.tsx'
import { CalendarPage } from './pages/CalendarPage.tsx'
import { CategoriesPage } from './pages/CategoriesPage.tsx'
import { DashboardPage } from './pages/DashboardPage.tsx'
import { ImportPage } from './pages/ImportPage.tsx'
import { ItemCardPage } from './pages/ItemCardPage.tsx'
import { ItemFormPage } from './pages/ItemFormPage.tsx'
import { ItemsPage } from './pages/ItemsPage.tsx'
import { LoginPage } from './pages/LoginPage.tsx'
import { NotificationsPage } from './pages/NotificationsPage.tsx'
import { ProfilePage } from './pages/ProfilePage.tsx'

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route element={<GuestRoute />}>
          <Route path="/login" element={<LoginPage />} />
        </Route>
        <Route element={<ProtectedRoute />}>
          <Route element={<Layout />}>
            <Route path="/" element={<DashboardPage />} />
            <Route path="/items" element={<ItemsPage />} />
            <Route element={<AdminRoute />}>
              <Route path="/items/new" element={<ItemFormPage />} />
              <Route path="/items/:id/edit" element={<ItemFormPage />} />
              <Route path="/import" element={<ImportPage />} />
              <Route path="/audit" element={<AuditPage />} />
            </Route>
            <Route path="/items/:id" element={<ItemCardPage />} />
            <Route path="/calendar" element={<CalendarPage />} />
            <Route path="/categories" element={<CategoriesPage />} />
            <Route path="/notifications" element={<NotificationsPage />} />
            <Route path="/profile" element={<ProfilePage />} />
          </Route>
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}
