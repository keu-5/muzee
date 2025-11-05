export const Header = () => {
  return (
    <header className="w-full h-16 border-b border-border bg-background sticky top-0 z-10 sm:hidden flex items-center justify-between px-4">
      <div className="flex items-center">
        <div className="w-9 h-9 rounded-full bg-muted flex items-center justify-center">
          <span className="text-sm text-muted-foreground">👤</span>
        </div>
      </div>

      <div className="absolute left-1/2 -translate-x-1/2">
        <p className="text-lg font-semibold tracking-wide">Muzee</p>
      </div>

      <div className="w-9" />
    </header>
  );
};
