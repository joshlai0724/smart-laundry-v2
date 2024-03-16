import { Card, List, Typography } from 'antd'
import { PageContent } from '@/components'
import styles from './index.module.scss'
import { useEffect } from 'react'
// import { storeApi } from '@/api/storeApi';
// import { storeUserApi } from '@/api/storeUserApi'

const data =[{
  type: 'type',
	created_by_user_id: 'created_by_user_id',
	created_by_user_name: 'created_by_user_name',
	user_id: 'user_id',
	user_name: 'user_name',
	device_id: 'device_id',
	device_name: 'device_name',
	device_real_type: 'device_real_type',
	device_display_type: 'device_display_type',
	amount: 64,
	point_amount: 100,
	ts: 210
},
{
  type: 'type1',
	created_by_user_id: 'created_by_user_id1',
	created_by_user_name: 'created_by_user_name1',
	user_id: 'user_id1',
	user_name: 'user_name1',
	device_id: 'device_id1',
	device_name: 'device_name1',
	device_real_type: 'device_real_type1',
	device_display_type: 'device_display_type1',
	amount: 641,
	point_amount: 1010,
	ts: 2110
}
]
// 顧客使用
const UsageRecords = () => {
  // const [loading, setLoading] = useState(false)
  // const [data, setData] = useState([])

  useEffect(() => {
    

  }, [])

  return (
    <PageContent title="儲值記錄">
            <Card  className={styles.listBox}>
          {/* <Card > */}
          <List
            // bordered
            itemLayout="horizontal"
            // loading={loading}
            header={<div>儲值記錄</div>}
            // itemLayout="vertical"
            size="small"
            dataSource={data}
            renderItem={(item: {
              type: string
              created_by_user_id: string
              created_by_user_name: string
              user_id: string
              user_name: string
              device_id: string
              device_name: string
              device_real_type: string
              device_display_type: string
              amount: number
              point_amount: number
              ts: number
            }) => (
              <List.Item >
                <List.Item.Meta
                  avatar={<Typography.Text mark>{item.type}</Typography.Text> }// 儲值
                  title={item.device_name}
                  description={item.ts}
                />
                <div>$ {item.amount}</div>
              </List.Item>
            )}
          />
        </Card>

    </PageContent>
  )
}

export default UsageRecords;
