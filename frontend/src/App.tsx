import { RouterProvider } from 'react-router-dom'
import router from './router'
import { ConfigProvider } from 'antd'
// import { useGlobalStore } from './stores'
import theme from '@/style/theme.json'
import 'antd/dist/reset.css'
import { Toaster } from 'react-hot-toast'

const App = () => {
  // const { primaryColor } = useGlobalStore()

  return (<ConfigProvider
    theme={theme}
  >    
    <RouterProvider router={router} />
    <Toaster />
  </ConfigProvider>)
}

export default App
