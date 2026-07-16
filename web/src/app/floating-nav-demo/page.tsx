import FloatingNav from '@/components/ui/floating-nav';

export default function FloatingNavDemo() {
  return (
    <div className="min-h-screen bg-tg-bg text-tg-text">
      <div className="px-4 pb-24 pt-6">
        <h1 className="text-2xl font-bold">FloatingNav Demo</h1>
        <p className="mt-2 text-sm text-tg-hint">
          Компонент плавающей нижней навигации. Используйте его как отдельную навигацию или
          замените им стандартный таб-бар на нужных экранах.
        </p>
      </div>
      <FloatingNav />
    </div>
  );
}
