import { jwtDecode } from "@/lib/jwt";
import { LINK } from "@/lib/links";
import axios, {
  AxiosError,
  AxiosRequestConfig,
  InternalAxiosRequestConfig,
} from "axios";

const isServer = typeof window === "undefined";

let isRefreshing = false;
let refreshQueue: (() => void)[] = [];

const instance = axios.create({
  baseURL: isServer ? "http://backend:8080" : "/api",
  withCredentials: true,
});

const HAS_PROFILE_KEY = "has_profile";

const saveHasProfile = (hasProfile: boolean) => {
  if (!isServer) {
    localStorage.setItem(HAS_PROFILE_KEY, String(hasProfile));
  }
};

const getHasProfile = (): boolean | null => {
  if (!isServer) {
    const value = localStorage.getItem(HAS_PROFILE_KEY);
    if (value === null) return null;
    return value === "true";
  }
  return null;
};

const clearHasProfile = () => {
  if (!isServer) {
    localStorage.removeItem(HAS_PROFILE_KEY);
  }
};

instance.interceptors.request.use((request) => {
  if (!isServer) {
    const hasProfile = getHasProfile();

    if (
      hasProfile === false &&
      !window.location.pathname.startsWith(LINK.createProfile)
    ) {
      window.location.replace(LINK.createProfile);
      return Promise.reject({
        message: "Redirecting to create profile",
        __REDIRECT__: true,
      });
    }
  }
  return request;
});

instance.interceptors.response.use(
  (response) => {
    if (!isServer) {
      try {
        const accessToken = response.data?.access_token;

        if (accessToken) {
          const decoded = jwtDecode(accessToken);
          saveHasProfile(decoded.has_profile);

          if (
            !decoded.has_profile &&
            !window.location.pathname.startsWith(LINK.createProfile)
          ) {
            window.location.replace(LINK.createProfile);
            return Promise.reject({
              message: "Redirecting to create profile",
              __REDIRECT__: true,
            });
          }
        }
      } catch (e) {
        console.warn("Token decode failed", e);
      }
    }
    return response;
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & {
      _retry?: boolean;
    };

    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      // リフレッシュエンドポイント自体の401は処理しない
      if (originalRequest.url?.includes("/auth/refresh")) {
        if (!isServer) {
          document.cookie = "access_token=; Max-Age=0; path=/";
          document.cookie = "refresh_token=; Max-Age=0; path=/";
          clearHasProfile();
          window.location.href = LINK.login;
        }
        return Promise.reject(error);
      }

      // ログインエンドポイントの401は処理しない（ログイン失敗はそのまま返す）
      if (originalRequest.url?.includes("/auth/login")) {
        return Promise.reject(error);
      }

      // 既にリフレッシュ中の場合は、キューに追加
      if (isRefreshing) {
        return new Promise((resolve) => {
          refreshQueue.push(() => {
            resolve(instance(originalRequest));
          });
        });
      }

      isRefreshing = true;

      try {
        // インターセプターを持たない専用のインスタンスでリフレッシュ
        const refreshInstance = axios.create({
          baseURL: isServer ? "http://backend:8080" : "/api",
          withCredentials: true,
        });

        // リフレッシュAPIを呼び出し
        await refreshInstance.post("/v1/auth/refresh", {
          refresh_token: undefined,
          client_id: process.env.NEXT_PUBLIC_CLIENT_ID || "",
        });

        // 新しいトークンが Cookie に設定された
        // キューに溜まっているリクエストを再実行
        refreshQueue.forEach((callback) => callback());
        refreshQueue = [];

        // 元のリクエストを再実行
        return instance(originalRequest);
      } catch (refreshError) {
        // リフレッシュ失敗時は Cookie をクリアしてログインページへ
        refreshQueue = [];
        if (!isServer) {
          document.cookie = "access_token=; Max-Age=0; path=/";
          document.cookie = "refresh_token=; Max-Age=0; path=/";
          clearHasProfile();
          window.location.href = LINK.login;
        }
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  },
);

const customAxios = async <T = unknown>(
  config: AxiosRequestConfig,
  options?: AxiosRequestConfig,
): Promise<T> => {
  try {
    const res = await instance.request({
      ...config,
      ...options,
      headers: {
        ...config.headers,
        ...options?.headers,
      },
    });

    return res.data;
  } catch (err) {
    const error = err as AxiosError<{ error?: string; message?: string }>;

    if (error.response?.data?.message) {
      (error as any).message = error.response.data.message;
    }

    if (isServer) {
      console.error("Failed to request:", {
        url: config.url,
        method: config.method,
        status: error.response?.status,
        data: error.response?.data,
      });
    }

    throw error;
  }
};

export default customAxios;
