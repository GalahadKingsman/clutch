import type { User } from '../lib/api';

type Props = { user: User };

export function ProfilePage({ user }: Props) {
  return (
    <div className="px-4 pt-6">
      <h1 className="font-display text-xl font-bold">Профиль</h1>
      <div className="mt-6 flex items-center gap-4">
        <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-blue/20 text-2xl font-bold">
          {user.first_name[0]}
        </div>
        <div>
          <p className="text-lg font-bold">
            {user.first_name} {user.last_name || ''}
          </p>
          {user.telegram_username && (
            <p className="text-sm text-mut">@{user.telegram_username}</p>
          )}
        </div>
      </div>
      <p className="mt-6 text-sm text-mut">
        Рейтинг, XP и честь — Phase 2.
      </p>
    </div>
  );
}
