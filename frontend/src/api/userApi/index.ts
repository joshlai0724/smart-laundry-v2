import { httpClient } from "@/utils/request";
import { LoginParams, PasswordParams } from "./types";
const BaseUrl = import.meta.env.VITE_BACKEND_HOST;
// const BaseUrl = window.location.protocol + "//" + window.location.hostname + "/api/v1/";

export const userApi = {
  login: async (params: LoginParams) => {
    return await httpClient.post<any>(BaseUrl + "users/login", params);
  },
  info: async () => {
    return await httpClient.get(BaseUrl + "users/info");
  },
  scope: async () => {
    return await httpClient.get(BaseUrl + "users/scopes");
  },
  resetPassword: async (params: PasswordParams) => {
    return await httpClient.post<any>(
      BaseUrl + "users/.change-password",
      params
    );
  },
  updateUsername: async (name: any) => {
    // console.log(name)
    return await httpClient.post<any>(BaseUrl + "users/update-self-info", name);
  },
};
