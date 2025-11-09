export const LINK = {
  createProfile: "/create-profile",
  home: "/home",
  login: "/login",
  logout: "/logout",
  signup: {
    base: "/signup",
    verify: "/signup/verify",
  },
};

export const PUBLIC_LINK = [LINK.login, LINK.signup.base, LINK.signup.verify];

export const generateStaticLink = (path: string | undefined): string | null => {
  if (!path) {
    return null;
  }

  return (
    process.env.NEXT_PUBLIC_S3_PUBLIC_URL +
    "/" +
    process.env.NEXT_PUBLIC_S3_PUBLIC_BUCKET +
    "/" +
    path
  );
};
