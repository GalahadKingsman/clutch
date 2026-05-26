import { Component, type ErrorInfo, type ReactNode } from 'react';

type Props = { children: ReactNode; fallbackTitle?: string };
type State = { error: Error | null };

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error): State {
    return { error };
  }

  componentDidCatch(error: Error, info: ErrorInfo) {
    console.error('CLUTCH UI error:', error, info.componentStack);
  }

  render() {
    if (this.state.error) {
      return (
        <div className="flex min-h-screen flex-col items-center justify-center px-6 text-center">
          <p className="font-display text-lg font-bold text-red">
            {this.props.fallbackTitle ?? 'Ошибка интерфейса'}
          </p>
          <p className="mt-3 text-sm text-mut">{this.state.error.message}</p>
          <button
            type="button"
            className="mt-6 rounded-xl bg-blue px-4 py-2 text-sm font-bold"
            onClick={() => window.location.reload()}
          >
            Перезагрузить
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}
