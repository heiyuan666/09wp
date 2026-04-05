import { BrowserRouter, Navigate, Route, Routes, useLocation } from 'react-router-dom';
import { Home } from './pages/Home';
import { Login } from './pages/Login';
import { AdminDashboard } from './pages/AdminDashboard';
import { Dashboard } from './pages/Dashboard';
import { Settings } from './pages/Settings';
import { ScheduledTasks } from './pages/ScheduledTasks';
import { ProtectedRoute } from './components/ProtectedRoute';
import { AuthProvider } from './contexts/AuthContext';
import { Toaster } from './components/ui/sonner';
import { Footer } from './components/Footer';

function AppLayout() {
  const location = useLocation();
  const showFooter = location.pathname === '/';

  return (
    <div className="min-h-screen flex flex-col">
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/admin/login" element={<Login />} />
        <Route
          path="/admin"
          element={
            <ProtectedRoute>
              <AdminDashboard />
            </ProtectedRoute>
          }
        >
          <Route index element={<Navigate to="dashboard" replace />} />
          <Route path="dashboard" element={<Dashboard />} />
          <Route path="settings" element={<Settings />} />
          <Route path="scheduled-tasks" element={<ScheduledTasks />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
      {showFooter && <Footer />}
    </div>
  );
}

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <AppLayout />
      </BrowserRouter>
      <Toaster />
    </AuthProvider>
  );
}

export default App;
