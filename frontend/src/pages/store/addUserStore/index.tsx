import { Card } from 'antd'
import { PageContent } from '@/components'
import styles from './index.module.scss'
// import { useEffect, useState } from 'react'
// import { storeApi } from '@/api/storeApi'
// import { storeUserApi } from '@/api/storeUserApi'
// import useUserInfoStore from '@/stores/userInfo'
// import { useNavigate, useParams } from 'react-router-dom'

const AddUserStore = () => {
  // const navigate = useNavigate()
  // const [messageApi] = message.useMessage()

  // // const [loading, setLoading] = useState(false)
  // const [data, setData] = useState()
  // const { token } = useUserInfoStore()
  // const store_id = useParams()

  // useEffect(() => {
  //   if (token && store_id) {
  //     navigate('/admin/storelists')
  //     storeUserApi.registerStore(store_id.store_id).then(() => {
  //       messageApi.open({
  //         type: 'success',
  //         content: '成功註冊分店'
  //       })
  //       navigate('/admin/storelists')
  //     })
  //   } else {
  //     messageApi.open({
  //       type: 'error',
  //       content: '註冊失敗'
  //     })
  //     // navigate("/login")
  //   }
  //   // setData(store);
  //   storeApi.getStores().then((response) => {
  //     setData(response.data.stores)
  //   })
  // }, [])

  return (
    <PageContent title="分店記錄123">
      <Card className={styles.listBox}>
      </Card>
    </PageContent>
  )
}

export default AddUserStore
