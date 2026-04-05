import { Link, Outlet, useLocation } from 'react-router-dom';
import { Button } from '@/components/ui/button';
import { useAuth } from '@/contexts/AuthContext';
import { Settings, LayoutDashboard, LogOut, Clock } from 'lucide-react';

export function AdminDashboard() {
  const { logout } = useAuth();
  const location = useLocation();

  const handleLogout = () => {
    logout();
    window.location.href = '/admin/login';
  };

  return (
    <div className="flex-1 bg-background flex flex-col">
      <nav className="border-b">
        <div className="container mx-auto px-4 py-4 flex items-center justify-between">
          <div className="flex items-center gap-6">
            <Link
              to="/"
              className="flex items-center gap-2 hover:opacity-80 transition-opacity cursor-pointer"
            >
              <img
                src="/images/logo.png"
                alt="PanCheck Logo"
                className="h-8 w-8"
              />
              <span className="text-lg font-semibold">PanCheck</span>
            </Link>
            <div className="flex gap-2">
              <Button
                variant={location.pathname === '/admin/dashboard' ? 'default' : 'ghost'}
                asChild
              >
                <Link to="/admin/dashboard">
                  <LayoutDashboard className="mr-2 h-4 w-4" />
                  仪表盘
                </Link>
              </Button>
            <Button
              variant={location.pathname === '/admin/scheduled-tasks' ? 'default' : 'ghost'}
              asChild
            >
              <Link to="/admin/scheduled-tasks">
                <Clock className="mr-2 h-4 w-4" />
                定时任务
              </Link>
            </Button>
            <Button
              variant={location.pathname === '/admin/settings' ? 'default' : 'ghost'}
              asChild
            >
              <Link to="/admin/settings">
                <Settings className="mr-2 h-4 w-4" />
                配置
              </Link>
            </Button>
            </div>
          </div>
          <Button variant="ghost" onClick={handleLogout}>
            <LogOut className="mr-2 h-4 w-4" />
            退出登录
          </Button>
        </div>
      </nav>
      <div className="container mx-auto py-8">
        <Outlet />
      </div>
    </div>
  );
}

