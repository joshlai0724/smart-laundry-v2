import { Card, List, Space, Tabs, type TabsProps, Row, Col, Button, Flex, Modal, Drawer, Form, InputNumber } from 'antd'
import { PageContent } from '@/components'
import styles from './index.module.scss'
import { useEffect, useState } from 'react'
// import { storeApi } from '@/api/storeApi';
import { storeUserApi } from '@/api/storeUserApi'
import { storeDeviceApi } from '@/api/storeDevice'
import { useUserInfoStore } from '@/stores'
import Icon from "@mdi/react";
import { mdiTumbleDryer } from "@mdi/js";
import toast from 'react-hot-toast'

// 顧客使用
const MachineList = () => {
  // interface IBalance {
  //   points: number
  //   balance: number
  // }
  // const [loading, setLoading] = useState(false);
  const [washMachines, setWashMachines] = useState([])
  const [dryMachines, setDryMachines] = useState([])
  const [banlance, setBalance] = useState<any>([])
  const [deviceId, setDeviceId] = useState('')
  const { currentStore } = useUserInfoStore()
  const [modal, contextHolder] = Modal.useModal()

  useEffect(() => {
    if(currentStore){
      console.log('currentStore');
      storeUserApi.balance(currentStore, 'c3d4daba-01b8-45fb-9926-333ab1cb115e').then((response) => {
        // console.log(response.data);
        setBalance(response.data);
      })
      storeDeviceApi.getStoreDevices(currentStore).then((response) => {
        // console.log(response.data)
        setWashMachines(response.data.devices.filter((res: any) => res.display_type === "washer"))
        setDryMachines(response.data.devices.filter((res: any) => res.display_type === "dryer"))
        // setLoading(false);
      })
    }

  }, [])

  const showState = (id: string) => {
    storeDeviceApi.getDevicesStatus(currentStore, id).then((response) => {
      // console.log(response.data)
      modal.success({
        title: '資訊',
        content: `狀態 ： ${response.data.state} 點數 ： ${response.data.points}`
      })
    })
  }

  // set insert coins drawer
  const [openDrawer, setOpenDrawer] = useState(false)

  const showDrawer = (deviceId: string) => {
    setDeviceId(deviceId);
    setOpenDrawer(true)
  }

  const insertCoins = (values: any) => {
    console.log(values);
    console.log(deviceId)
    const insertCoin = storeDeviceApi.insertCoins(currentStore, deviceId, values);
    toast.promise(insertCoin,{
      loading: 'Loading',
      success: '投幣成功',
      error: (err) => `投幣失敗: ${err.toString()}`
    })
    insertCoin.then(() => {
      setOpenDrawer(false)
    }, (error) => {
      const resMessage =
        (error.response?.data?.message) ||
        error.message ||
        error.toString()
      console.log(resMessage)
    })
  }
  // set insert coins drawer end

  // set top-up drawer
  const [openTopupDrawer, setOpenTopupDrawer] = useState(false)

  const showTopupDrawer = () => {
    // setDeviceId(deviceId);
    setOpenTopupDrawer(true)
  }

  const topUp = (values: any) => {
    const topupData = storeDeviceApi.insertCoins(currentStore, deviceId, values)
    toast.promise(topupData,{
      loading: 'Loading',
      success: '儲值成功',
      error: (err) => `儲值失敗: ${err.toString()}`
    })
    topupData.then(() => {
      setOpenTopupDrawer(false)
    }, (error) => {
      const resMessage =
        (error.response?.data?.message) ||
        error.message ||
        error.toString()
      console.log(resMessage)
    })
  }
  // set top-up end
  const items: TabsProps['items'] = [
    {
      key: '1',
      label: '洗衣機',
      children: (<>

        <Card className={styles.listBox}>
          {/* <Card > */}
          <List
            // bordered
            // itemLayout="horizontal"
            // loading={loading}
            itemLayout="vertical"
            size="small"
            dataSource={washMachines}
            renderItem={(item: { real_type: string, id: string, name: string, display_type: string }) => (
              <List.Item >
                <Row>
                  <Col xs={24} sm={8} md={8} lg={8} xl={8} xxl={8}>
                    <List.Item.Meta
                      avatar={<Icon path={mdiTumbleDryer} size={2} />}
                      title={item.display_type == 'washer' ? item.name + ' 洗衣機' : item.name + ' 烘衣機'}
                    // description={item.real_type}
                    />
                  </Col>
                  <Col xs={24} sm={16} md={16} lg={16} xl={16} xxl={16}>
                    <Flex justify="flex-end" gap="small" wrap="wrap" >
                      <Button onClick={() => { showState(item.id) }}>取得狀態</Button>
                      <Button onClick={() => { showDrawer(item.id) }}>投幣</Button>
                    </Flex>
                  </Col>
                </Row>
                <Space>
                  {/* <Button onClick={() => showUpdateModal(item.id, item.name)}>編輯</Button> */}
                </Space>
              </List.Item>
            )}
          />
        </Card>
      </>)
    },
    {
      key: '2',
      label: '烘衣機',
      children: (<>

        <Card className={styles.listBox}>
          {/* <Card > */}
          <List
            // bordered
            // itemLayout="horizontal"
            // loading={loading}
            itemLayout="vertical"
            size="small"
            dataSource={dryMachines}
            renderItem={(item: { real_type: string, id: string, name: string, display_type: string }) => (
              <List.Item >
                <Row>
                  <Col xs={24} sm={8} md={8} lg={8} xl={8} xxl={8}>
                    <List.Item.Meta
                      avatar={<Icon path={mdiTumbleDryer} size={2} />}
                      title={item.display_type == 'washer' ? item.name + ' 洗衣機' : item.name + ' 烘衣機'}
                    // description={item.real_type}
                    />
                  </Col>
                  <Col xs={24} sm={16} md={16} lg={16} xl={16} xxl={16}>
                    <Flex justify="flex-end" gap="small" wrap="wrap" >
                      <Button onClick={() => { showState(item.id) }}>取得狀態</Button>
                      <Button onClick={() => { showDrawer(item.id) }}>投幣</Button>
                    </Flex>
                  </Col>
                </Row>
                <Space>
                  {/* <Button onClick={() => showUpdateModal(item.id, item.name)}>編輯</Button> */}
                </Space>
              </List.Item>
            )}
          />
        </Card>
      </>)
    }
  ]
  const info = () => {
    // messageApi.info('Hello, Ant Design!');
    toast.success('登入成功')
  };
  return (
    <PageContent title="洗衣烘衣列表">
      {contextHolder}
      <Drawer title={'投幣'} placement="bottom" onClose={() => { setOpenDrawer(false) }} open={openDrawer} height={240}>
        <Form layout="vertical" onFinish={insertCoins} initialValues={{ amount: 100 }}>
          <Form.Item
            name="amount"
            label="金額"
            rules={[{ required: true, message: '請輸入投幣金額' }]}
          >
            <InputNumber
              style={{ width: '100%' }}
              size='large'
              step={10} // 每次加減10
            // onChange={onChange}
            />
          </Form.Item>
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
            // className={styles.login_form_button}
            // disabled={loading}
            >
              儲值
            </Button>
          </Form.Item>
        </Form>
      </Drawer>
      <Drawer title={'儲值'} placement="bottom" onClose={() => { setOpenTopupDrawer(false) }} open={openTopupDrawer} height={240}>
        <Form layout="vertical" onFinish={topUp} initialValues={{ amount: 100 }}>
          <Form.Item
            name="amount"
            label="金額"
            rules={[{ required: true, message: '請輸入儲值金額' }]}
          >
            <InputNumber
              style={{ width: '100%' }}
              size='large'
              step={10} // 每次加減10
            // onChange={onChange}
            />
          </Form.Item>
          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
            // className={styles.login_form_button}
            // disabled={loading}
            >
              儲值
            </Button>
          </Form.Item>
        </Form>
      </Drawer>

      <Card>
        <Row >
          <Col flex={2}>
            餘額：{banlance.balance}
          </Col>
          <Col flex={2} >
            點數：{banlance.points}
          </Col>
          <Col flex={6} >
            <Button onClick={() => showTopupDrawer()}>儲值</Button>
            <Button
              onClick={info}
            >
              測試
            </Button>
          </Col>

        </Row>
      </Card>
      <Tabs defaultActiveKey="1" items={items} type="card" />
    </PageContent>
  )
}

export default MachineList
