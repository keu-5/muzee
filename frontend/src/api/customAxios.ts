import axios, { AxiosError, AxiosRequestConfig } from "axios";

const isServer = typeof window === "undefined";

const customAxios = async <T = unknown>(
  config: AxiosRequestConfig,
  options?: AxiosRequestConfig,
): Promise<T> => {
  const instance = axios.create({
    baseURL: isServer ? "http://backend:8080" : "/api",
  });

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
