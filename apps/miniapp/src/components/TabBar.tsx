import { NavLink } from 'react-router-dom';

const linkClass = ({ isActive }: { isActive: boolean }) =>
  `flex flex-col items-center gap-0.5 text-[9px] font-bold ${
    isActive ? 'text-blue' : 'text-[#6E6B80]'
  }`;

export function TabBar() {
  return (
    <nav className="fixed bottom-3 left-3 right-3 z-20 flex h-[60px] items-center justify-around rounded-[22px] border border-white/10 bg-[rgba(20,18,28,0.92)] px-2 backdrop-blur-md">
      <NavLink to="/feed" className={linkClass}>
        <span className="text-lg">📰</span>
        Лента
      </NavLink>
      <NavLink to="/friends" className={linkClass}>
        <span className="text-lg">👥</span>
        Друзья
      </NavLink>
      <NavLink
        to="/duel/create"
        className="-mt-4 flex h-12 w-12 flex-col items-center justify-center rounded-[15px] bg-gradient-to-b from-[#5C88FF] to-[#4068E8] text-2xl text-white shadow-[0_6px_18px_rgba(77,124,255,0.45)]"
      >
        +
      </NavLink>
      <NavLink to="/wallet" className={linkClass}>
        <span className="text-lg">💰</span>
        Кошелёк
      </NavLink>
      <NavLink to="/profile" className={linkClass}>
        <span className="text-lg">👤</span>
        Профиль
      </NavLink>
    </nav>
  );
}
