import toast from "react-hot-toast";
import { create } from "zustand";
import { createJSONStorage, persist } from "zustand/middleware";

interface UserInfoProps {
  name: string;
}

interface Scopes {
  // 不確定寫法
  role: string;
  // "user:data:write": boolean;
  // "user:data:read": boolean;
  // "store:read": boolean;
  // "store:create": boolean;
  // "store:enable": boolean;
  // "store:deactive": boolean;
  // "store:write": boolean;
  // "store:password:write": boolean;
  // "store:report:read": boolean;
  // "store:user:admin:register": boolean;
  // "store:user:hq:register": boolean;
  // "store:user:cust:register": boolean;
}

interface UserInfoState {
  isLogin: boolean;
  token: string;
  refreshToken: string;
  userInfo: UserInfoProps | null;
  userScopes: Scopes[] | null;
  storeScopes: Scopes[] | null;
  currentStore: string;
  setUserInfo: (value: UserInfoProps) => void;
  setUserScopes: (value: Scopes[]) => void;
  setStoreScopes: (value: Scopes[]) => void;
  setToken: (token: string) => void;
  setRefreshToken: (refreshToken: string) => void;
  setCurrentStore: (currentStore: string) => void;
  logout: () => void;
}
/**
 * token
 * 使用者資料
 * storage
 */
const useUserInfoStore = create<UserInfoState>()(
  persist(
    (set) => ({
      isLogin: false,
      token: "",
      refreshToken: "",
      userInfo: null,
      userScopes: null,
      storeScopes: null,
      currentStore: "",
      setUserInfo: (userInfo: UserInfoProps) => {
        set(() => ({ userInfo }));
      },
      setUserScopes: (userScopes: Scopes[]) => {
        set(() => ({ userScopes }));
      },
      setStoreScopes: (storeScopes: Scopes[]) => {
        set(() => ({ storeScopes }));
      },
      setToken: (token: string) => {
        set(() => ({ token, isLogin: true }));
      },
      setRefreshToken: (refreshToken: string) => {
        set(() => ({ refreshToken, isLogin: true }));
      },
      setCurrentStore: (currentStore: string) => {
        set(() => ({ currentStore }));
      },
      logout: () => {
        toast.success('登出');
        set(() => ({
          userInfo: null,
          userScopes: null,
          isLogin: false,
          token: "",
          refreshToken: "",
          currentStore: "",
        }));
      },
    }),
    {
      name: "USER_INFO",
      storage: createJSONStorage(() => localStorage),
    }
  )
);

export default useUserInfoStore;
