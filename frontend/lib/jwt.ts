interface AccessTokenPayload {
  user_id: number;
  email: string;
  has_profile: boolean;
  exp: number;
  iat: number;
}

export const jwtDecode = (token: string): AccessTokenPayload => {
  const payload = token.split(".")[1];
  return JSON.parse(atob(payload)) as AccessTokenPayload;
};
