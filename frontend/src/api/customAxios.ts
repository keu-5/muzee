import { LINK } from "@/lib/links";
import axios, {
  AxiosError,
  AxiosRequestConfig,
  InternalAxiosRequestConfig,
} from "axios";

const isServer = typeof window === "undefined";

let isRefreshing = false;
let refreshQueue: (() => void)[] = [];

const customAxios = async <T = unknown>(
  config: AxiosRequestConfig,
  options?: AxiosRequestConfig,
): Promise<T> => {
  const instance = axios.create({
    baseURL: isServer ? "http://backend:8080" : "/api",
    withCredentials: true,
  });

  // レスポンスインターセプター：401エラー時にリフレッシュ
  instance.interceptors.response.use(
    (response) => response,
    async (error: AxiosError) => {
      const originalRequest = error.config as InternalAxiosRequestConfig & {
        _retry?: boolean;
      };

      // 401エラーで、まだリトライしていない場合
      if (error.response?.status === 401 && !originalRequest._retry) {
        originalRequest._retry = true;

        // リフレッシュエンドポイント自体の401は処理しない
        if (originalRequest.url?.includes("/auth/refresh")) {
          if (!isServer) {
            document.cookie = "access_token=; Max-Age=0; path=/";
            document.cookie = "refresh_token=; Max-Age=0; path=/";
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
