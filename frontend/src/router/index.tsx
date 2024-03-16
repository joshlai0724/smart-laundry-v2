import { lazy } from 'react'
import { createBrowserRouter } from 'react-router-dom'
import type { RouteObject } from 'react-router-dom'
import { lazyLoad } from './util'
import Layout from '@/components/layout'
import { BarsOutlined, ContainerOutlined } from '@ant-design/icons'
import TopupRecords from '@/pages/user/topupRecords'
import UsageRecords from '@/pages/user/usageRecords'
// import { BarsOutlined, ContainerOutlined } from '@ant-design/icons'

// const authLoader = () => {
//   const token = useUserInfoStore.getState().userInfo?.token

//   if (!token) {
//     return redirect(`/login?to=${window.location.pathname + window.location.search}`)
//   }

//   return null
// }
const Login = lazy(async () => await import('@/pages/loginForm'))
const Register = lazy(async () => await import('@/pages/registerForm'))
const ForgotPassword = lazy(async () => await import('@/pages/forgotPassword/index'))
const SetForgotPassword = lazy(async () => await import('@/pages/setForgotPassword/index'))
const ErrorBoundary = lazy(async () => await import('@/components/error-boundary'))
// const EditStore = lazy(() => import('@/pages/admin/store/storeEdit'));
const MachineList = lazy(async () => await import('@/pages/device/machineList'))
const StoreList = lazy(async () => await import('@/pages/store/storeList'))
const StoreUsersList = lazy(async () => await import('@/pages/store/storeUsers'))
const AddUserStore = lazy(async () => await import('@/pages/store/addUserStore'))
const StoreMachines = lazy(async () => await import('@/pages/store/storeMachines'))
const StoreRecords = lazy(async () => await import('@/pages/store/storeRecords'))
// const TopupRecords = lazy(async () => await import('@/pages/user/topupRecords'))
// const UsageRecords = lazy(async () => await import('@/pages/user/usageRecords'))

// admin路由(包含hq和admin)
export const adminRoutes = [
  {
    index: true,
    title: '首頁',
    icon: 'IconHome',
    element: <MachineList />
  },
  // {
  //   path: 'home',
  //   title: '首頁',
  //   icon: 'IconHome',
  //   element: <Home />,
  // },
  {
    path: 'Machine',
    title: '洗衣/烘衣',
    icon: 'IconHome',
    element: <MachineList />
    // userScope: '123',
  },
  { path: 'topuprecords', icon: <BarsOutlined />, title: '儲值記錄', element: lazyLoad(<TopupRecords />) },
  { path: 'usagerecords', icon: <ContainerOutlined />, title: '使用記錄', element: lazyLoad(<UsageRecords />) },
  {
    type: 'group',
    path: 'admin',
    title: '........',
    children: [
      {
        path: 'storelists',
        title: '店鋪列表',
        icon: 'LayoutOutlined',
        userScope: 'store:create',
        element: lazyLoad(<StoreList />)
        // children: [
        //   ,
        // ],
      },
      {
        path: 'addstore/:store_id',
        title: '新增分店',
        userScope: 'store:user:cust:register',
        isMenu: false,
        element: lazyLoad(<AddUserStore />)
      },
      {
        path: 'storemachines',
        title: '分店機器列表',
        // userScope: 'store:user:cust:register',
        isMenu: false,
        element: lazyLoad(<StoreMachines />)
      },
      { path: 'storeuserlist', title: '分店使用者', isMenu: false, element: lazyLoad(<StoreUsersList />) },
      { path: 'storerecords', title: '記錄查詢', isMenu: false, element: lazyLoad(<StoreRecords />) },
      // {
      //   path: 'form',
      //   title: '紀錄查詢',
      //   icon: 'FormOutlined',
      //   children: [
      //     { path: 'base', title: '儲值記錄', element: lazyLoad(<StoreUsersList />) },
      //     { path: 'base', title: '使用記錄', element: lazyLoad(<StoreUsersList />) },
      //   ],
      // },
    ]
  }
  // loggedIn ? admin :{} // scope
]

const routes: RouteObject[] = [
  {
    path: 'login',
    element: <Login />
  },
  {
    path: 'register',
    // title: '註冊',
    element: lazyLoad(<Register />)
  },
  {
    path: 'forgetpassword',
    // title: '忘記密碼',
    element: lazyLoad(<ForgotPassword />)
  },
  {
    path: 'setforgetpassword/:verCode',
    element: <SetForgotPassword />
  },
  {
    path: '/',
    element: <Layout />,
    errorElement: <ErrorBoundary />,
    children: [
      ...adminRoutes
      // {
      //   path: 'storeuserlist',
      //   // loader: () => ({ isAuth: false, ac: 'ac' }),
      //   element: lazyLoad(<StoreUsersList />),
      // },

      // {
      //   path: 'hotnews',
      //   element: lazyLoad(lazy(() => import('@/pages/HotNews'))),
      // },
    ]
  }
]

const router = createBrowserRouter(routes, {
  basename: import.meta.env.BASE_URL
})

export default router
