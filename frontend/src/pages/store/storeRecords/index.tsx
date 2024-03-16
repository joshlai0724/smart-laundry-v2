import { Button, Card, List, Flex,  Row, Col, Tabs, type TabsProps } from 'antd'
import { PageContent } from '@/components'
import styles from './index.module.scss'
import { useLocation } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { storeUserApi } from '@/api/storeUserApi'
// import useUserInfo from '@/stores/userInfo'

// const records = [// 200
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   },
//   {

//     type: 'st123ring',
//     user_id: 'st123ring',
//     user_name: 'st123ring',
//     device_id: 'st123ring',
//     device_name: 'st123ring',
//     device_type: 'string1',
//     from_user_id: 'string2',
//     from_user_name: 'string3',
//     from_online_payment: 'string5',
//     amount: 123,
//     point_amount: 456,
//     ts: 123

//   }

// ]

const StoreRecords = () => {
  // const [loading, setLoading] = useState(false)
  const [data, setData] = useState()
  // const { userScopes } = useUserInfo()
  const { state } = useLocation()
  // console.log(state.storeId);
  useEffect(() => {
    // setData(records)
    storeUserApi.getStoreUsers(state.storeId).then((response) => {
      // console.log(response.data)
      setData(response.data.users)
    })
  }, [])

  // set top-up end
  const items: TabsProps['items'] = [
    {
      key: '1',
      label: '使用記錄',
      children: (
        <div>
          <span>-----------------</span>
        </div>
      )
    },
    {
      key: '2',
      label: '儲值記錄',
      children: (<>

<Card className={styles.listBox}>
        <List
          bordered
          itemLayout="horizontal"
          // itemLayout="vertical"
          // className={styles.listBox}
          size="small"
          dataSource={data}
          renderItem={(item: { user_name: string, from_online_payment: string, device_name: string, state: string, role: string }) => (
            <List.Item>
                  {/* <List.Item.Meta
          title={<a href="https://ant.design">{item.user_name}</a>}
          description="Ant Design, a design language for background applications, is refined by Ant UED Team"
        /> */}
          <Row >
                <Col xs={24} sm={8} md={8} lg={8} xl={8} xxl={8}>
                    <Button>{item.device_name}</Button>
                    <Button>{item.user_name}</Button>
                </Col>
                <Col xs={24} sm={16} md={16} lg={16} xl={16} xxl={16}>
                  <Flex justify="flex-end" gap="small" wrap="wrap" >
                  <Button>{item.user_name}</Button>
                  </Flex>
                </Col>
              </Row>
            </List.Item>
          )}
        />
      </Card>

      </>)
    }
  ]

  return (
    <PageContent title="分店使用者" back>
      <Tabs defaultActiveKey="1" items={items} type="card" className={styles.listBox} />

    </PageContent>
  )
}

export default StoreRecords
