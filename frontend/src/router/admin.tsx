import StoreList from '@/pages/store/storeList/'
// import { lazy } from 'react'
import { lazyLoad } from './util'

// 表单
// const UserInfo = lazy(async () => await import('@/pages/form/user'))
// userApi.scope().then((response) => {console.log(response.data.scopes)})

// const DateTimeForm = lazy(() => import('@/pages/admin/form/datetime'));
// const LinkageForm = lazy(() => import('@/pages/admin/form/linkage'));
// const CustomForm = lazy(() => import('@/pages/admin/form/custom'));
// const FilterForm = lazy(() => import('@/pages/admin/form/filter'));
// const UploadForm = lazy(() => import('@/pages/admin/form/upload'));
// const ModalFormPage = lazy(() => import('@/pages/admin/form/modal'));

export default {
  type: 'group',
  path: 'admin',
  title: 'admin',
  children: [
    {
      path: 'layout',
      title: '店鋪設定',
      icon: 'LayoutOutlined',
      children: [
        { path: 'storelists', title: '店鋪列表', element: lazyLoad(<StoreList />) }
        // { path: 'Lists', title: 'Lists', element: lazyLoad(<Lists />) },
      ]
    },
    {
      path: 'form',
      title: 'user',
      icon: 'FormOutlined',
      children: [
        // { path: 'base', title: '使用者資料修改', element: lazyLoad(<UserInfo />) }
        // { path: 'datetime', title: '日期时间', element: withLoadingComponent(<DateTimeForm />) },
        // { path: 'linkage', title: '表单联动', element: withLoadingComponent(<LinkageForm />) },
        // { path: 'upload', title: '上传', element: withLoadingComponent(<UploadForm />) },
        // { path: 'custom', title: '自定义表单', element: withLoadingComponent(<CustomForm />) },
        // { path: 'filter', title: '筛选表单', element: withLoadingComponent(<FilterForm />) },
        // { path: 'modal', title: '弹窗表单', element: withLoadingComponent(<ModalFormPage />) },
      ]
    }
  ]
}
