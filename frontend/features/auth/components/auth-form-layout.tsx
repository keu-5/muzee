import Image from "next/image";

interface AuthPageLayoutProps {
  children: React.ReactNode;
  title: string;
  description: string;
}

export const AuthFormLayout = ({
  children,
  title,
  description,
}: AuthPageLayoutProps) => {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-background via-background to-primary/5 p-4 relative overflow-hidden">
      <div className="w-full max-w-md space-y-8">
        <div className="flex flex-col items-center space-y-2">
          <div className="relative w-16 h-16 mb-2">
            <Image
              src="/muzee-logo.png"
              alt="Logo"
              width={64}
              height={64}
              className="rounded-xl"
            />
          </div>
          <h1 className="text-3xl font-bold text-center">{title}</h1>
          <p className="text-muted-foreground text-center">{description}</p>
        </div>

        {children}
      </div>
    </div>
  );
};
