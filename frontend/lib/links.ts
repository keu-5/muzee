export const LINK = {
  createProfile: "/create-profile",
  home: "/home",
  login: "/login",
  signup: {
    base: "/signup",
    verify: "/signup/verify",
  },
};

export const PUBLIC_LINKS = [LINK.login, LINK.signup.base, LINK.signup.verify];
