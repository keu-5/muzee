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

//TODO: 本番環境対応
export const generateStaticLink = (path: string | undefined): string | null => {
  if (!path) {
    return null;
  }

  return "http://localhost/storage" + "/" + "public-uploads" + "/" + path;
};
