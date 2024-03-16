import { Button, Card, List, Popconfirm, Form, Modal, Flex, Row, Col, Drawer, InputNumber } from 'antd'
import { PageContent } from '@/components'
import styles from './index.module.scss'
import { Link, useLocation } from 'react-router-dom'
import { UserOutlined } from '@ant-design/icons'
import { useEffect, useState } from 'react'
import { storeUserApi } from '@/api/storeUserApi'
import toast from 'react-hot-toast'

const StoreUsersList = () => {
  // const [loading, setLoading] = useState(false);
  const [data, setData] = useState()
  // const { userScopes } = useUserInfo();
  const { state } = useLocation()
  const [modal, contextHolder] = Modal.useModal()
  console.log(state.storeId)
  useEffect(() => {
    // setData(users);
    storeUserApi.getStoreUsers(state.storeId).then((response) => {
      // console.log(response.data)
      setData(response.data.users)
    })
  }, [])

  const showBalance = (id: string) => {
    storeUserApi.balance(state.storeId, id).then((response) => {
      console.log(response.data)
      modal.success({
        title: '使用者資訊',
        content: `餘額 ： ${response.data.balance} 點數 ： ${response.data.points}`
      })
    })
  }

  // 設定user權限modal start
  const [open, setOpen] = useState(false)
  const [selectedName, setSelectedName] = useState('')
  const [selectedId, setSelectedId] = useState('')
  const [selectedRole, setSelectedRole] = useState('')
  const showUpdateModal = (id: string, name: string, role: string) => {
    setOpen(true)
    setSelectedName(name)
    setSelectedId(id)
    setSelectedRole(role)
  }
  // const onUpdateUser = (values: any) => {
  //   storeApi.updateStore(values, selectedId)
  //     .then(() => {
  //       storeApi.getStores().then((response) => {
  //         console.log(response.data.stores)
  //         setData(response.data.stores)
  //       })
  //     }, (error) => {
  //       const resMessage =
  //         (error.response?.data?.message) ||
  //         error.message ||
  //         error.toString()
  //       console.log(resMessage)
  //     }
  //     )
  //   console.log('Received values of form: ', values)
  //   setOpen(false)
  // }
  // set user modal end

  // 啟用禁用user
  const enableUser = (id: string) => {
    const enableUserData = storeUserApi.enableUser(state.storeId, id).then(() => {
      storeUserApi.getStoreUsers(state.storeId).then((response) => {
        console.log(response.data.users)
        setData(response.data.users)
      })
    })
    toast.promise(enableUserData,{
      loading: 'Loading',
      success: '更新成功',
      // success: '成功登入',
      error: (err) => `更新失敗: ${err.toString()}`
    })
  }

  const deactiveUser = (id: string) => {
    const deactiveUserData =  storeUserApi.deactiveUser(state.storeId, id).then(() => {
      storeUserApi.getStoreUsers(state.storeId).then((response) => {
        console.log(response.data.users)
        setData(response.data.users)
      })
    })
    toast.promise(deactiveUserData,{
      loading: 'Loading',
      success: '更新成功',
      // success: '成功登入',
      error: (err) => `更新失敗: ${err.toString()}`
    })
  }

  // set top-up drawer
  const [openDrawer, setOpenDrawer] = useState(false)

  const showDrawer = (userId: string, userName: string) => {
    setSelectedId(userId)
    setSelectedName(userName)
    setOpenDrawer(true)
  }
  // set top-up end

  const topUp = (values: any) => {
    const cashTopupData = storeUserApi.cashTopup(values, state.storeId, selectedId).then(() => {
      setOpenDrawer(false)
    }, (error) => {
      const resMessage =
        (error.response?.data?.message) ||
        error.message ||
        error.toString()
      console.log(resMessage)
    })
    toast.promise(cashTopupData,{
      loading: 'Loading',
      success: '現金儲值成功',
      // success: '成功登入',
      error: (err) => `現金儲值 失敗: ${err.toString()}`
    })
  }

  const  changeMgr = (storeId: string,selected: string) =>{
    const changeData =  storeUserApi.changetoMgr(storeId, selected);
    toast.promise(changeData,{
      loading: 'Loading',
      success: '已變更為店長',
      // success: '成功登入',
      error: (err) => `變更失敗: ${err.toString()}`
    })
  }

  const  changeOwner = (storeId: string,selected: string) =>{
    const changeData =  storeUserApi.changetoOwner(storeId, selected);
    toast.promise(changeData,{
      loading: 'Loading',
      success: '已變更為業主',
      // success: '成功登入',
      error: (err) => `變更失敗: ${err.toString()}`
    })
  }

  const  changeCust = (storeId: string,selected: string) =>{
    const changeData =  storeUserApi.changetoOwner(storeId, selected);
    toast.promise(changeData,{
      loading: 'Loading',
      success: '已變更為消費者',
      // success: '成功登入',
      error: (err) => `變更失敗: ${err.toString()}`
    })
  }
  return (
    <PageContent title="分店使用者" back>
      {contextHolder}
      <Card className={styles.listBox}>
        <Modal title="編輯權限" open={open} onCancel={() => { setOpen(false) }}>
          <Flex wrap="wrap" gap="small">
            {selectedRole !== 'cust' && <Popconfirm title="確認設定為消費者?" onConfirm={() => changeCust(state.store_id, selectedId)} okText="確認" cancelText="取消">
              <Button danger type="dashed">設定為消費者</Button>
            </Popconfirm>}
            {selectedRole !== 'mgr' && <Popconfirm title="確認設定為店長?" onConfirm={ () => changeMgr(state.store_id, selectedId)} okText="確認" cancelText="取消">
              <Button danger type="dashed">設定為店長</Button>
            </Popconfirm>}
            {selectedRole !== 'owner' && <Popconfirm title="確認設定為業主?" onConfirm={async () =>  changeOwner(state.store_id, selectedId)} okText="確認" cancelText="取消">
              <Button danger type="dashed">設定為業主</Button>
            </Popconfirm>}
            {/* {selectedRole !=="hq"&&<Popconfirm title="確認設定為總部?" onConfirm={() => storeUserApi.changetoHq(store_id,selectedId)} okText="確認" cancelText="取消">
          <Button danger type="dashed">設定為總部</Button>
        </Popconfirm>} */}
            {/* {selectedRole !=="admin"&&<Popconfirm title="確認設定為管理者?" onConfirm={() => deactiveUser(selectedId)} okText="確認" cancelText="取消">
          <Button danger type="dashed">設定為管理者</Button>
        </Popconfirm>} */}
          </Flex>
        </Modal>
        <Drawer title={selectedName + '現金儲值'} placement="bottom" onClose={() => { setOpenDrawer(false) }} open={openDrawer} height={240}>
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

        <List
          // bordered
          // itemLayout="horizontal"
          itemLayout="vertical"
          className={styles.listBox}
          dataSource={data}
          renderItem={(item: { phone_number: string, id: string, name: string, state: string, role: string }) => (
            <List.Item className={styles.listBox}>
              <Row>
                <Col xs={24} sm={8} md={8} lg={8} xl={8} xxl={8}>
                  <List.Item.Meta
                    avatar={<UserOutlined />}
                    title={item.name}
                    description={item.phone_number + item.role}
                  />
                </Col>
                <Col xs={24} sm={16} md={16} lg={16} xl={16} xxl={16}>
                  <Flex justify="flex-end" gap="small" wrap="wrap" >

                    <Button onClick={() => { showUpdateModal(item.id, item.name, item.role) }}>編輯權限</Button>
                    {item.state == 'active' && <Popconfirm title="確認禁用?" onConfirm={() => { deactiveUser(item.id) }} okText="確認" cancelText="取消"><Button danger type="dashed"
                    >禁用</Button></Popconfirm>}
                    {item.state != 'active' && <Popconfirm title="確認啟用?" onConfirm={() => { enableUser(item.id) }} okText="確認" cancelText="取消"><Button type="primary"
                    >啟用</Button></Popconfirm>}
                    <Button onClick={() => { showBalance(item.id) }}>取得餘額</Button>
                    <Button onClick={() => { showDrawer(item.id, item.name) }}>現金儲值</Button>
                    <Button> <Link to="/admin/storerecords" state= {{ storeId: item.id }}>紀錄查詢</Link></Button>
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

export default StoreUsersList
