import { Link, useNavigate } from 'react-router-dom'
import { Button, Form, Input } from 'antd'
import { LockOutlined, UserOutlined } from '@ant-design/icons'
import styles from './index.module.scss'
import useUserInfo from '@/stores/userInfo'
import { useState } from 'react'
import toast  from 'react-hot-toast';

import { userApi } from '@/api/userApi'

const Login = () => {
  const navigate = useNavigate()
  const { setRefreshToken, setToken, setUserScopes } = useUserInfo()
  const [loading, setLoading] = useState<boolean>(false);

  interface FieldType {
    phone_number?: string
    password?: string
  }

  const onFinish = async (values: FieldType) => {
    // console.log("Success:", values);

    // const { phone_number, password } = values;
    const loginData =  userApi.login(values);
    loginData.then((response) => {
        // console.log(response);
        if (response.data.access_token) {
          // console.log(response.data)
          setToken(response.data.access_token)
          setRefreshToken(response.data.refresh_token)
          userApi.scope().then(async (response) => {
            // console.log(response.data.scopes);
            await setUserScopes(response.data.scopes)
            navigate('/');
            setLoading(false);
          })
        }
        // navigate(0);
      }, (error) => {
        const resMessage =
          (error.response?.data?.message) ||
          error.message ||
          error.toString()
          navigate('/')
        setLoading(false);
        console.log(resMessage)
      }
      )
      toast.promise(loginData,
        {
        loading: 'Loading',
        success: '成功登入',
        // success: '成功登入',
        error: (err) => `登入失敗: ${err.toString()}`
      },{
        success: {
          duration: 5000,
          icon: '🔥',
        },
      }
      )
  }

  return (
    <div className={styles.login_page}>
      <Form
        name="normal_login"
        className={styles.login_form}
        // initialValues={{ remember: true }}
        onFinish={onFinish}
      >
        <Form.Item<FieldType>
          name="phone_number"
          rules={[{ required: true, message: '請輸入手機號碼!' }]}
        >
          <Input prefix={<UserOutlined />} placeholder="手機號碼" />
        </Form.Item>
        <Form.Item<FieldType>
          name="password"
          rules={[{ required: true, message: '請輸入密碼!' }]}
        >
          <Input prefix={<LockOutlined />} type="password" placeholder="密碼" />
        </Form.Item>
        <Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            className={styles.login_form_button}
            loading={loading}
          >
            登入
          </Button>
        </Form.Item>
        <Form.Item>
          <Link to="/register"> 註冊</Link>
          <Link style={{ float: 'right' }} to="/forgetpassword"> 忘記密碼</Link>
        </Form.Item>
      </Form>
    </div>
  )
}

export default Login
