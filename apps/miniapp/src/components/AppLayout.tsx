import { Outlet } from 'react-router-dom';
import { TabBar } from './TabBar';

export function AppLayout() {
  return (
    <div className="min-h-screen pb-24">
      <Outlet />
      <TabBar />
    </div>
  );
}
