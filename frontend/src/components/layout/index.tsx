import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { MenuUnfoldOutlined, MenuFoldOutlined, LogoutOutlined, SettingOutlined } from '@ant-design/icons'
import { Layout, Menu, Dropdown, Space, Button, type MenuProps, Form, Input, Modal, Select } from 'antd'
import { ErrorBoundary } from '@/components'
import { adminRoutes } from '@/router'
import { getAllPath } from '@/router/util'
import useUserInfoStore from '@/stores/userInfo'
import { icons } from './config'
import styles from './index.module.scss'
import { userApi } from '@/api/userApi'
import useUserInfo from '@/stores/userInfo'
import { storeApi } from '@/api/storeApi'
import { storeUserApi } from '@/api/storeUserApi'

const { Sider, Content } = Layout
interface Values {
  title: string
  description: string
  modifier: string
}

/* ResetPasswordModal Start*/
interface ResetPasswordFormProps {
  open: boolean
  onResetPassword: (values: Values) => void
  onCancel: () => void
}

const ResetPasswordForm: React.FC<ResetPasswordFormProps> = ({
  open,
  onResetPassword,
  onCancel
}) => {
  const [form] = Form.useForm()
  return (
    <Modal
      open={open}
      title="修改密碼"
      okText="修改"
      cancelText="取消"
      onCancel={onCancel}
      onOk={() => {
        form
          .validateFields()
          .then((values) => {
            form.resetFields()
            onResetPassword(values)
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
        initialValues={{ modifier: 'public' }}
      >
        <Form.Item
          name="old_password"
          label="舊密碼"
          rules={[
            {
              required: true,
              message: '請輸入密碼!'
            }
          ]}
          hasFeedback
        >
          <Input.Password />
        </Form.Item>
        <Form.Item
          name="new_password"
          label="新密碼"
          rules={[
            {
              required: true,
              message: '請輸入密碼!'
            }
          ]}
          hasFeedback
        >
          <Input.Password />
        </Form.Item>

        <Form.Item
          name="confirm"
          label="確認新密碼"
          dependencies={['new_password']}
          hasFeedback
          rules={[
            {
              required: true,
              message: '請輸入密碼!'
            },
            ({ getFieldValue }) => ({
              async validator(_, value) {
                if (!value || getFieldValue('new_password') === value) {
                  await Promise.resolve(); return
                }
                return await Promise.reject(new Error('新密碼不一樣！！'))
              }
            })
          ]}
        >
          <Input.Password />
        </Form.Item>
      </Form>
    </Modal>
  )
}
/* ResetPasswordModal End*/
/* UpdateUsernameModal */
interface UpdateUsernameFormProps {
  open2: boolean
  onUpdateUsername: (values: Values) => void
  onCancel: () => void
}

const UpdateUsernameForm: React.FC<UpdateUsernameFormProps> = ({
  open2,
  onUpdateUsername,
  onCancel
}) => {
  const [updateinfoform] = Form.useForm()
  return (
    <Modal
      open={open2}
      title="修改密碼"
      okText="修改"
      cancelText="取消"
      onCancel={onCancel}
      onOk={() => {
        updateinfoform
          .validateFields()
          .then((values) => {
            updateinfoform.resetFields()
            onUpdateUsername(values)
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
        initialValues={{ modifier: 'public' }}
      >
        <Form.Item
          name="name"
        >
          <Input.Password />
        </Form.Item>
      </Form>
    </Modal>
  )
}
/* UpdateUsernameModal End*/

/* changeStoreModal */
interface ChangeStoreProps {
  openStore: boolean
  currentStore: boolean
  option: any
  onChangeStore: (values: Values) => void
  onCancel: () => void
}

const ChangeStore: React.FC<ChangeStoreProps> = ({
  openStore,
  currentStore,
  option,
  onChangeStore,
  onCancel
}) => {
  const [changestoreform] = Form.useForm()
  // console.log(currentStore);
  return (
    <Modal
      open={openStore}
      title="選擇分店"
      okText="確認"
      cancelText="取消"
      onCancel={onCancel}
      maskClosable={false}
      onOk={() => {
        changestoreform
          .validateFields()
          .then((values) => {
            changestoreform.resetFields()
            onChangeStore(values)
          })
          .catch((info) => {
            console.log('Validate Failed:', info)
          })
      }}
    >
      <Form
        form={changestoreform}
        layout="vertical"
        name="form_in_modal"
        initialValues={{ modifier: 'public' }}
      >
        {currentStore && <div className="form-group">
          <div className="alert alert-danger" role="alert">
            第一次登入請先選擇分店
          </div>
        </div>}
        <Form.Item
          name="id"

        >
          <Select
            // style={{ width:  }}
            options={option.map((res: any) => ({ label: res.name, value: res.id }))}
          />
        </Form.Item>
      </Form>
    </Modal>
  )
}
/* UpdateUsernameModal End*/

const AppLayout = () => {
  const { isLogin, logout, token, currentStore, setCurrentStore } = useUserInfoStore()

  const location = useLocation()
  const nav = useNavigate()
  const { userInfo, userScopes, setUserInfo, setUserScopes } = useUserInfo()
  const [selStore, setSelStore] = useState([])
  const [userStore, setUserStore] = useState([])
  const [hadStore, sethadStore] = useState(false)

  const [collapsed, setCollapsed] = useState(false)
  const [selectedKeys, setSelectedKeys] = useState<string[]>([])

  const openKeys = getAllPath(location.pathname)

  const handleClick = ({ key }: any) => {
    nav(key)
  }

  const toggleCollapsed = () => {
    setCollapsed(!collapsed)
  }
  // 啟用網頁先確認登入狀態
  useEffect(() => {
    // const user = getCurrentUser()
    // const user = localStorage.getItem("USER_INFO")
    // console.log(token)
    // 如果有token抓取name和scopes
    if (token) {
      userApi.info().then((response) => {
        // // console.log(token);
        setUserInfo(response.data)
      })
      userApi.scope().then((response) => {
        // // console.log(response.data.scopes);
        setUserScopes(response.data.scopes)
      })
      // setCurrentStore('123');
      // console.log(currentStore)
      if (!currentStore) {
        sethadStore(true);
        // console.log(hadStore)
      }
      storeApi.getStores().then((response) => {
        // console.log(response.data.stores)
        setSelStore(response.data.stores)
      })
      storeUserApi.getUserStores().then(async (response) => {
        // setOption(response.data.stores)
        console.log(response.data.stores)
        setUserStore(response.data.stores) // 確認使用者的註冊名單
        if (response.data.stores.length === 0) {
          setOpenStore(true) // 使用者沒有分店需先註冊
        }else{
          if(!currentStore){
            console.log('loadStore')
            await setCurrentStore(response.data.stores[0].id);
            // nav('/', { replace: true })
            nav(0);
          }
        }

      })

    }
    if (!isLogin) {
      nav('/login')
    }
  }, [isLogin, nav])

  // scopes之後會寫在這邊
  const treeForeach = useCallback((tree: any, path?: any) => {
    return tree
      ?.map((data: any) => {
        // const { index, isMenu, title, path: _path, children, icon, ...other } = data;
        const { index, isMenu, userScope, title, path: _path, children, icon, ...other } = data
        const pathArr = path ? [...path, _path] : [_path]
        if (isMenu === false) {
          return false
        }
        // // console.log(userScope)
        if (!userScopes?.includes(userScope) && typeof (userScope) !== 'undefined') {
          return false
        }
        if (index) {
          return false
        }
        if (icon && typeof icon === 'string') {
          other.icon = React.createElement(icons[icon])
        } else if (icon === '' || icon === undefined) {
          other.icon = undefined
        } else {
          other.icon = icon
        }
        return {
          ...other,
          label: title,
          path: _path,
          key: `/${pathArr.join('/')}`,
          children: children ? treeForeach(children, pathArr) : undefined
        }
      })
      .filter(Boolean)
  }, [])

  const antdMenuTree = useMemo(() => {
    return treeForeach(adminRoutes)
  }, [treeForeach])

  useEffect(() => {
    const { pathname } = location
    setSelectedKeys(getAllPath(pathname))
  }, [location])

  // 右上menu
  const items: MenuProps['items'] = [
    {
      key: 'user',
      label: userInfo?.name,
      children: [
        {
          label: '設定密碼',
          key: 'resetpassword',
          onClick: () => {
            setOpen(true)
          }
        },
        {
          label: '修改使用者名稱',
          key: 'updateusename',
          onClick: () => {
            setOpen2(true)
          }
        }
      ]
    },
    {
      // 更換分店，之後以modal設定
      key: 'changestore',
      label: <span>更換分店</span>,
      // icon: <LogoutOutlined />,
      onClick: () => { setOpenStore(true) }
    },
    {
      key: 'logout',
      label: <span>登出</span>,
      icon: <LogoutOutlined />,
      onClick: () => { logout() }
    }
  ]

  // reset password
  const [open, setOpen] = useState(false)

  const onResetPassword = (values: any) => {
    userApi.resetPassword(values)
      .then((response) => {
        console.log(response)
      }, (error) => {
        const resMessage =
          (error.response?.data?.message) ||
          error.message ||
          error.toString()
        console.log(resMessage)
      }
      )
    // console.log('Received values of form: ', values)
    setOpen(false)
  }

  // update user name
  const [open2, setOpen2] = useState(false)

  const onUpdateUsername = (values: any) => {
    userApi.updateUsername(values)
      .then((response) => {
        console.log(response)
      }, (error) => {
        const resMessage =
          (error.response?.data?.message) ||
          error.message ||
          error.toString()
        console.log(resMessage)
      }
      )
    // console.log('Received values of form: ', values)
    setOpen2(false)
  }

  // changeStore
  const [openStore, setOpenStore] = useState(false)
  interface res{
    id:string;
  }
  const onChangeStore = (values: any) => {
    if(userStore.find((res :res) => res.id ===values.id)){
      setCurrentStore(values.id);
      setOpenStore(false)
      nav(0); // 刷新頁面重設store資料
    }
    else{
      storeUserApi.registerStore(values.id).then(() =>{
        setCurrentStore(values.id);
        nav(0);// 刷新頁面重設store資料
      })
    }
    // console.log('change store to: ', values);
    setOpenStore(false);
  }

  return (
    <div className={styles.layout}>
      <div className={styles.header}>
        <div>
          {/* <img className={styles.logo} src={logoImg} alt="logo" /> */}
          <Button onClick={toggleCollapsed} className={styles.btn}>
            {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
          </Button>

        </div >
        <div className={styles.rightArea}>
          <div className={styles.userInfo}>
            <Dropdown menu={{ items }} placement="bottomRight">
              <a onClick={(e) => { e.preventDefault() }}>
                <Space>
                  <SettingOutlined />
                </Space>
              </a>

            </Dropdown>
          </div>
        </div>
      </div >
      <Layout className={styles.content}>
        <div className={styles.leftArea}>
          <div className={styles.siderBox}>
            {antdMenuTree.length > 0 && (
              <Sider
                width={220}
                trigger={null}
                collapsible
                collapsed={collapsed}
                onCollapse={(value: any) => { setCollapsed(value) }}
              >
                <Menu
                  className={styles.subMenu}
                  mode="inline"
                  selectedKeys={selectedKeys}
                  defaultOpenKeys={openKeys}
                  items={antdMenuTree}
                  onClick={handleClick}
                />
              </Sider>
            )}
          </div>
          <div className={styles.customCollapsed}>
            <Button onClick={toggleCollapsed} className={styles.btn}>
              {collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            </Button>
          </div>
        </div>
        <Content className={styles.mainContent} id="mainContent">
          <ResetPasswordForm
            open={open}
            onResetPassword={onResetPassword}
            onCancel={() => {
              setOpen(false)
            }}
          />
          <UpdateUsernameForm
            open2={open2}
            onUpdateUsername={onUpdateUsername}
            onCancel={() => {
              setOpen2(false)
            }}
          />
          <ChangeStore
            openStore={openStore}
            currentStore={hadStore}
            option={selStore}
            onChangeStore={onChangeStore}
            onCancel={() => {
              setOpenStore(false)
            }}
          />
          <ErrorBoundary>
            <Outlet />
          </ErrorBoundary>
        </Content>
      </Layout>
    </div >
  )
}

export default AppLayout
