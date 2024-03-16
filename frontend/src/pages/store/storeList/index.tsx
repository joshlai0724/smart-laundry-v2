import { Button, Card, List, Space, Popconfirm, Input, Form, Modal, Col, Row, Flex } from 'antd'
import { PageContent } from '@/components'
import styles from './index.module.scss'
import { Link } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { storeApi } from '@/api/storeApi'
import toast from 'react-hot-toast'

interface AddStoreFormProps {
  open: boolean
  onCreateStore: (values: any) => void
  onCancel: () => void
}

const AddStoreForm: React.FC<AddStoreFormProps> = ({
  open,
  onCreateStore,
  onCancel
}) => {
  const [form] = Form.useForm()
  return (
    <Modal
      open={open}
      title="新增店鋪"
      okText="新增"
      cancelText="取消"
      onCancel={onCancel}
      onOk={() => {
        form
          .validateFields()
          .then((values) => {
            form.resetFields()
            onCreateStore(values)
          })
          .catch((info) => {
            console.log('Validate Failed:', info)
          })
      }}
    >
      <Form
        form={form}
        layout="vertical"
        name="form_in_modal"
      >
        <Form.Item
          name="name"
          label="店鋪名稱"
          rules={[
            {
              required: true,
              message: '請輸入店鋪名稱!'
            }
          ]}
          hasFeedback
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="address"
          label="地址"
          rules={[
            {
              required: true,
              message: '請輸入地址!'
            }
          ]}
          hasFeedback
        >
          <Input />
        </Form.Item>
      </Form>
    </Modal>
  )
}

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
          label="新店鋪名稱"
          rules={[
            {
              required: true,
              message: '請輸入新帳號!'
            }
          ]}
          hasFeedback
        >
          <Input />
        </Form.Item>
        <Form.Item
          name="address"
          label="新店鋪地址"
          rules={[
            {
              required: true,
              message: '請輸入新地址!'
            }
          ]}
          hasFeedback
        >
          <Input />
        </Form.Item>
      </Form>
    </Modal>
  )
}

const StoreList = () => {
  // const [loading, setLoading] = useState(false)
  const [data, setData] = useState()

  useEffect(() => {
    // setData(store);
    storeApi.getStores().then((response) => {
      setData(response.data.stores)
    })
  }, [])

  // add new store
  const [open, setOpen] = useState(false)

  const onCreateStore = (values: any) => {
    const createStoreData = storeApi.createStore(values)
      toast.promise(createStoreData,{
        loading: 'Loading',
        success: '建立成功',
        // success: '成功登入',
        error: (err) => `建立失敗: ${err.toString()}`
      })
      createStoreData.then((response) => {
        console.log(response)
        storeApi.getStores().then((response) => {
          console.log(response.data.stores)
          setData(response.data.stores)
        })
      }, (error) => {
        const resMessage =
          (error.response?.data?.message) ||
          error.message ||
          error.toString()
        console.log(resMessage)
      }
      )
    // console.log('Received values of form: ', values);
  }

  // update user name
  const [open2, setOpen2] = useState(false)
  const [selectedName, setSelectedName] = useState('')
  const [selectedId, setSelectedId] = useState('')
  const showUpdateModal = (id: string, name: string) => {
    setOpen2(true)
    setSelectedName(name)
    setSelectedId(id)
  }
  const onUpdateStore = (values: any) => {
    const updateStoreData = storeApi.updateStore(values, selectedId)
    toast.promise(updateStoreData,{
      loading: 'Loading',
      success: '更新成功',
      // success: '成功登入',
      error: (err) => `更新失敗: ${err.toString()}`
    })
    updateStoreData.then(() => {
        storeApi.getStores().then((response) => {
          console.log(response.data.stores)
          setData(response.data.stores)
        })
      }, (error) => {
        const resMessage =
          (error.response?.data?.message) ||
          error.message ||
          error.toString()
        console.log(resMessage)
      }
      )
      
    // console.log('Received values of form: ', values);
    setOpen2(false)
  }

  const enableStore = (id: string) => {
    const enableStore = storeApi.enableStore(id).then(() => {
      storeApi.getStores().then((response) => {
        console.log(response.data.stores)
        setData(response.data.stores)
      })
    })
    toast.promise(enableStore,{
      loading: 'Loading',
      success: '更新成功',
      // success: '成功登入',
      error: (err) => `更新失敗: ${err.toString()}`
    })
  }

  const deactiveStore = (id: string) => {
    const deactiveStoreData = storeApi.deactiveStore(id).then(() => {
      storeApi.getStores().then((response) => {
        console.log(response.data.stores)
        setData(response.data.stores)
      })
    })
    toast.promise(deactiveStoreData,{
      loading: 'Loading',
      success: '更新成功',
      // success: '成功登入',
      error: (err) => `更新失敗: ${err.toString()}`
    })
  }
  // console.log(data)
  return (
    <PageContent title="店家list">
      <AddStoreForm
        open={open}
        onCreateStore={onCreateStore}
        onCancel={() => {
          setOpen(false)
        }}
      />
      <UpdateStoreForm
        open2={open2}
        selData={selectedName}
        onUpdateStore={onUpdateStore}
        onCancel={() => {
          setOpen2(false)
        }}
      />
      <Card>
        <Space>
          <Button type="primary" onClick={() => { setOpen(true) }} >新增店鋪</Button>
        </Space>
      </Card>
      <Card className={styles.listBox}>
        {/* <Card > */}
        <List
          // bordered
          // itemLayout="horizontal"
          itemLayout="vertical"
          className={styles.listBox}
          dataSource={data}
          renderItem={(item: { address: string, id: string, name: string, state: string }) => (
            <List.Item className={styles.listBox}>
              <Row>
                <Col xs={24} sm={8} md={8} lg={8} xl={8} xxl={8}>
                  <List.Item.Meta
                    //  avatar={<CheckCircleOutlined />}
                    title={item.state == 'active' ? item.name + ' 開店中' : item.name + ' 關店中'}
                    description={item.address}
                  />
                </Col>
                <Col xs={24} sm={16} md={16} lg={16} xl={16} xxl={16}>
                <Flex justify="flex-end" gap="small" wrap="wrap" >
                      <Button onClick={() => { showUpdateModal(item.id, item.name) }}>編輯</Button>
                      <Button >產生密碼</Button>
                      {/* {item.state=='enable' ? '開店中': '關店中'} */}
                      {item.state == 'active' && <Popconfirm title="確認關店?" onConfirm={() => { deactiveStore(item.id) }} okText="確認" cancelText="取消"><Button danger type="dashed"
                      >關店</Button></Popconfirm>}
                      {item.state != 'active' && <Popconfirm title="確認開店?" onConfirm={() => { enableStore(item.id) }} okText="確認" cancelText="取消"><Button type="primary"
                      >開店</Button></Popconfirm>}
                      <Button ><Link to="/admin/storemachines" state= {{ storeId: item.id }}>分店機器列表</Link></Button>
                      <Button ><Link to="/admin/storeuserlist" state= {{ storeId: item.id }}>使用者列表</Link></Button>
                      <Button><Link to="">交易記錄</Link></Button>
                  </Flex>
                </Col>
              </Row>
            </List.Item>
          )}
        />
      </Card>
    </PageContent>
  )
}

export default StoreList
