import { httpClient } from "@/utils/request";
import { createStoreParams } from "./types";
// const BaseUrl = import.meta.env.VITE_BACKEND_HOST;

const BaseUrl = window.location.protocol + "//" + window.location.hostname + "/api/v1/";

export const storeApi = {
  getStores: async () => {
    return await httpClient.get(BaseUrl + "stores");
  },
  getStore: async (store_id: string) => {
    return await httpClient.get(BaseUrl + "store/" + store_id);
  },
  createStore: async (params: createStoreParams) => {
    return await httpClient.post(BaseUrl + "stores/.create", params);
  },
  updateStore: async (params: createStoreParams, store_id: string) => {
    return await httpClient.post(
      BaseUrl + "stores/" + store_id + "/update-info",
      params
    );
  },
  enableStore: async (store_id: string) => {
    return await httpClient.post(BaseUrl + "stores/" + store_id + "/.enable");
  },
  deactiveStore: async (store_id: string) => {
    return await httpClient.post(BaseUrl + "stores/" + store_id + "/.deactive");
  },
};
