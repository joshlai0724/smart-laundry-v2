import axios from "axios";
import type {
  AxiosInstance,
  AxiosRequestConfig,
  AxiosResponse,
  CreateAxiosDefaults,
  InternalAxiosRequestConfig,
} from "axios";
import { useUserInfoStore } from "@/stores";

// 參考 https://blog.theashishmaurya.me/handling-jwt-access-and-refresh-token-using-axios-in-react-app
// 設定axios接口
declare module "axios" {
  interface AxiosRequestConfig {
    successCode?: number;
  }
}

interface ApiResult<T = unknown> {
  code: number;
  data: T;
  message: string;
}

// const SUCCESS_CODE = 204; // 空的成功回傳值

class Request {
  private readonly instance: AxiosInstance;

  private readonly abortControllerMap: Map<string, AbortController>;

  constructor(config: CreateAxiosDefaults) {
    this.instance = axios.create(config);

    this.abortControllerMap = new Map();

    // header加上token
    this.instance.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        if (config.url !== "/login") {
          const token = useUserInfoStore.getState().token; // 取得access_token
          if (token) {
            // config.headers["x-token"] = token;
            config.headers.Authorization = `Bearer ${token}`;
          }
        }

        const controller = new AbortController();
        const url = config.url || "";
        config.signal = controller.signal;
        this.abortControllerMap.set(url, controller);
        // console.log("test");
        return config;
      },
      async (error) => await Promise.reject(error)
    );

    // 回應攔截
    this.instance.interceptors.response.use(
      (response: AxiosResponse) => {
        const url = response.config.url || "";
        this.abortControllerMap.delete(url);
        return response;
      },
      async (err) => {
        /**
         * refresh token？
         */
        // 401/403 無token認證失敗先重新登入
        const originalRequest = err.config;
        if (err.response?.status === 401 || err.response?.status === 403) {
          try {
            const refreshToken = useUserInfoStore.getState().refreshToken; // 取得refresh_token
            const response = await axios.post("/users/renew-access-token", {
              refresh_token: refreshToken,
            });
            // console.log(response.data);
            useUserInfoStore.setState({ token: response.data.access_token });
            const token = useUserInfoStore.getState().token; // 取得refresh_token
            // console.log(token);
            originalRequest.headers.Authorization = `Bearer ${token}`;
            return await axios(originalRequest);
          } catch (error) {
            console.log(error);
            // logout();
            useUserInfoStore.setState({
              userInfo: null,
              userScopes: null,
              isLogin: false,
              token: "",
              refreshToken: "",
            });
            // window.location.href = `/web/login?redirect=${window.location.pathname}`;
          }
        }

        return await Promise.reject(err.message);
      }
    );
  }

  // cancelAllRequest() {
  //   for (const [, controller] of this.abortControllerMap) {
  //     controller.abort();
  //   }
  //   this.abortControllerMap.clear();
  // }

  // cancelRequest(url: string | string[]) {
  //   const urlList = Array.isArray(url) ? url : [url];
  //   for (const _url of urlList) {
  //     this.abortControllerMap.get(_url)?.abort();
  //     this.abortControllerMap.delete(_url);
  //   }
  // }

  async request<T = any>(config: AxiosRequestConfig): Promise<ApiResult<T>> {
    return await this.instance.request(config);
  }

  async get<T = any>(
    url: string,
    config?: AxiosRequestConfig
  ): Promise<ApiResult<T>> {
    // console.log(url);
    return await this.instance.get(url, config);
  }

  async post<T = any>(
    url: string,
    data?: any,
    config?: AxiosRequestConfig
  ): Promise<ApiResult<T>> {
    return await this.instance.post(url, data, config);
  }
}

export const httpClient = new Request({
  timeout: 20 * 1000,
});
