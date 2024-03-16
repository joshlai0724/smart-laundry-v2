import { Card, List, Space, Select, Tabs, type TabsProps, Row, Col, Button, Flex, Modal, Drawer, Form, InputNumber, Input } from 'antd'
import { PageContent } from '@/components'
import styles from './index.module.scss'
import { useEffect, useState } from 'react'
// import { storeApi } from '@/api/storeApi';
// import { storeUserApi } from '@/api/storeUserApi'
import { storeDeviceApi } from '@/api/storeDevice'
// import { useUserInfoStore } from '@/stores'
import Icon from "@mdi/react";
import { mdiTumbleDryer } from "@mdi/js";
import { useLocation } from 'react-router-dom'

interface UpdateStoreFormProps {
  open2: boolean
  selData: any
  onUpdateStore: (values: any) => void
  onCancel: () => void
}

const UpdateStoreForm: React.FC<UpdateStoreFormProps> = ({
  open2,
  selData,
  onUpdateStore,
  onCancel
}) => {
  const [updateinfoform] = Form.useForm()
  return (
    <Modal
      open={open2}
      title={'修改' + selData}
      okText="修改"
      cancelText="取消"
      onCancel={onCancel}
      onOk={() => {
        updateinfoform
          .validateFields()
          .then((values) => {
            updateinfoform.resetFields()
            onUpdateStore(values)
          })
          .catch((info) => {
            console.log('Validate Failed:', info)
          })
      }}
    >
      <Form
        form={updateinfoform}
        layout="vertical"
        name="form_in_modal"
      >
        <Form.Item
          name="name"
          label="新機器名稱"
          rules={[
            {
              required: true,
              message: '請輸入新機器名稱!'
            }
          ]}
          hasFeedback
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="display_type"
          label="新機器類別"
          rules={[
            {
              required: true,
              message: '請輸入新機器類別!'
            }
          ]}
          hasFeedback
        >
          <Select
            // style={{ width:  }}
            options={[
              { value: 'washer', label: '洗衣機' },
              { value: 'dryer', label: '烘衣機' },
            ]}
          />
        </Form.Item>
      </Form>
    </Modal>
  )
}

const StoreMachines = () => {
  // const [loading, setLoading] = useState(false);
  const [washMachines, setWashMachines] = useState([])
  const [dryMachines, setDryMachines] = useState([])
  // const [banlance, setBalance] = useState([])
  const [deviceId, setDeviceId] = useState('')
  // const { state.storeId } = useUserInfoStore()
  const [modal, contextHolder] = Modal.useModal()
  const { state } = useLocation()

  useEffect(() => {

    storeDeviceApi.getStoreDevices(state.storeId).then((response) => {
      console.log(response.data)
      setWashMachines(response.data.devices.filter((res: any) => res.display_type === "washer"))
      setDryMachines(response.data.devices.filter((res: any) => res.display_type === "dryer"))
      // setLoading(false);
    })
  }, [])

  const showState = (id: string) => {
    storeDeviceApi.getDevicesStatus(state.storeId, id).then((response) => {
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
    storeDeviceApi.insertCoins(state.storeId, deviceId, values).then(() => {
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

  const [open2, setOpen2] = useState(false)
  const [selectedName, setSelectedName] = useState('')
  const [selectedId, setSelectedId] = useState('')

  const showUpdateModal = (id: string, name: string) => {
    setOpen2(true)
    setSelectedName(name)
    setSelectedId(id)
  }
  const onUpdateStore = (values: any) => {
    storeDeviceApi.updateInfo(state.storeId, selectedId, values)
      .then(() => {
        storeDeviceApi.getStoreDevices(state.storeId).then((response) => {
          // console.log(response.data)
          setWashMachines(response.data.devices.filter((res: any) => res.display_type === "washer"))
          setDryMachines(response.data.devices.filter((res: any) => res.display_type === "dryer"))
          // setLoading(false);
        })
      }, (error) => {
        const resMessage =
          (error.response?.data?.message) ||
          error.message ||
          error.toString()
        console.log(resMessage)
      }
      )
    console.log('Received values of form: ', values);
    setOpen2(false)
  }

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
                      title={ item.name}
                    // description={item.real_type}
                    />
                  </Col>
                  <Col xs={24} sm={16} md={16} lg={16} xl={16} xxl={16}>
                    <Flex justify="flex-end" gap="small" wrap="wrap" >
                      <Button onClick={() => { showUpdateModal(item.id, item.name) }}>編輯</Button>
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
                      title={ item.name}
                    // description={item.real_type}
                    />
                  </Col>
                  <Col xs={24} sm={16} md={16} lg={16} xl={16} xxl={16}>
                    <Flex justify="flex-end" gap="small" wrap="wrap" >
                      <Button onClick={() => { showUpdateModal(item.id, item.name) }}>編輯</Button>
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

  return (
    <PageContent title="洗衣烘衣列表" back>
      {contextHolder}
      <UpdateStoreForm
        open2={open2}
        selData={selectedName}
        onUpdateStore={onUpdateStore}
        onCancel={() => {
          setOpen2(false)
        }}
      />
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
      <Tabs defaultActiveKey="1" items={items} type="card" />
    </PageContent>
  )
}

export default StoreMachines
