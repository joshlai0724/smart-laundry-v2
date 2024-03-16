import { httpClient } from "@/utils/request";
import { cashTopup } from "./types";
// const BaseUrl = import.meta.env.VITE_BACKEND_HOST;

const BaseUrl = window.location.protocol + "//" + window.location.hostname + "/api/v1/";

export const storeUserApi = {
  getUserStores: async () => {
    return await httpClient.get(BaseUrl + "/users/stores");
  },
  getStoreUsers: async (store_id: string) => {
    return await httpClient.get(BaseUrl + "stores/" + store_id + "/users");
  },
  // GET /api/v1/store/{store_id}/users/{user_id}/balance (scope: store:user:admin:records:read, store:user:hq:records:read, store:user:owner:records:read, store:user:mgr:records:read, store:user:cust:records:read, store:user:records:read)
  registerStore: async (store_id: string) => {
    return await httpClient.post(
      BaseUrl + "stores/" + store_id + "/users/.register"
    );
  },
  scopes: async (store_id: string) => {
    return await httpClient.post(
      BaseUrl + "stores/" + store_id + "/users/scopes"
    );
  },
  cashTopup: async (params: cashTopup, store_id: string, user_id: string) => {
    return await httpClient.post(
      BaseUrl +
        "stores/" +
        store_id +
        "/users/" +
        user_id +
        "/cust-cash-top-up",
      params
    );
  },
  balance: async (store_id: string, user_id: string) => {
    return await httpClient.get(
      BaseUrl + "stores/" + store_id + "/users/" + user_id + "/balance"
    );
  },
  enableUser: async (store_id: string, user_id: string) => {
    return await httpClient.post(
      BaseUrl + "stores/" + store_id + "/users/" + user_id + "/.enable"
    );
  },
  deactiveUser: async (store_id: string, user_id: string) => {
    return await httpClient.post(
      BaseUrl + "stores/" + store_id + "/users/" + user_id + "/.deactive"
    );
  },
  changetoOwner: async (store_id: string, user_id: string) => {
    return await httpClient.post(
      BaseUrl + "stores/" + store_id + "/users/" + user_id + "/change-to-owner"
    );
  },
  changetoMgr: async (store_id: string, user_id: string) => {
    return await httpClient.post(
      BaseUrl + "stores/" + store_id + "/users/" + user_id + "/change-to-mgr"
    );
  },
  changetoCust: async (store_id: string, user_id: string) => {
    return await httpClient.post(
      BaseUrl + "stores/" + store_id + "/users/" + user_id + "/change-to-cust"
    );
  },
};
