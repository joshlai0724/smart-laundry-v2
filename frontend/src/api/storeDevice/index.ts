import { httpClient } from "@/utils/request";
import { insertCoins } from "./types";
// const BaseUrl = import.meta.env.VITE_BACKEND_HOST;

const BaseUrl = window.location.protocol + "//" + window.location.hostname + "/api/v1/";
export const storeDeviceApi = {
  getStoreDevices: async (store_id: string) => {
    return await httpClient.get(BaseUrl + "stores/" + store_id + "/devices");
  },
  getDevicesStatus: async (store_id: string, device_id: string) => {
    return await httpClient.get(
      BaseUrl +
        "stores/" +
        store_id +
        "/coin-acceptors/" +
        device_id +
        "/status"
    );
  },
  insertCoins: async (
    store_id: string,
    device_id: string,
    params: insertCoins
  ) => {
    return await httpClient.post(
      BaseUrl +
        "stores/" +
        store_id +
        "/coin-acceptors/" +
        device_id +
        "/insert-coins",
      params
    );
  },
  updateInfo: async (
    store_id: string,
    device_id: string,
    params: insertCoins
  ) => {
    return await httpClient.post(
      BaseUrl +
        "stores/" +
        store_id +
        "/coin-acceptors/" +
        device_id +
        "/update-info",
      params
    );
  },
};
