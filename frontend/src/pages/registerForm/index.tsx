import { Link, useNavigate } from 'react-router-dom'
import { Button, Flex, Form, Input } from 'antd'
import { LockOutlined, UserOutlined } from '@ant-design/icons'
import styles from './index.module.scss'
// import { useState } from 'react'
import axios from 'axios'
import { useState } from 'react'
import toast from 'react-hot-toast'

const Register = () => {
  const navigate = useNavigate()
  const [loading, setLoading] = useState<boolean>(false)
  // const [message, setMessage] = useState<string>('')
  const [form] = Form.useForm()

  interface FieldType {
    phone_number: string
    password: string
    name: string
    ver_code: string
  }

  const onFinish = async (values: FieldType) => {
    console.log('Success:', values)
    const { phone_number, name, password, ver_code } = values
    setLoading(true)

    const registerData =axios.post('http://localhost:80/api/v1/users/.register', {
      phone_number,
      name,
      password,
      ver_code
    })
      toast.promise(registerData,{
        loading: 'Loading',
        success: '註冊成功',
        error: (err) => `註冊失敗: ${err.toString()}`
      })
      registerData.then((response) => {
        console.log(response)
        navigate('/login')
      }, (error) => {
        const resMessage =
          (error.response?.data?.message) ||
          error.message ||
          error.toString()
        setLoading(false)
        // setMessage(resMessage)
        console.log(resMessage)
      }
      )
  }

  // 取得註冊驗證碼
  const sendCheckMsg = () => {
    const phone = form.getFieldValue('phone_number')
    // console.log(phone);
    const checkMsg = axios.post('http://localhost:80/api/v1/users/' + 'send-check-phone-number-owner-msg', {
      phone_number: phone
    })
    toast.promise(checkMsg,{
      loading: 'Loading',
      success: '驗證碼已送出',
      error: (err) => `驗證碼送出失敗: ${err.toString()}`
    })
    checkMsg.then(
      () => {
        console.log('Success:', phone)
      },
      (error) => {
        const resMessage =
          (error.response?.data?.message) ||
          error.message ||
          error.toString()
        setLoading(false)
        console.log(resMessage)
      }
    )
  }
  return (
    <div className={styles.register_page}>
      <Form
        form={form}
        name="basic"
        className={styles.register_form}
        // initialValues={{ remember: true }}
        onFinish={onFinish}
      >
        <Form.Item
          name="phone_number"
          rules={[{ required: true, message: '請輸入手機號碼!' }]}
        >
          <Input prefix={<UserOutlined />} placeholder="手機號碼" />
        </Form.Item>
        <Form.Item
          name="name"
          rules={[{ required: true, message: '請輸入使用者名稱!' }]}
        >
          <Input prefix={<UserOutlined />} placeholder="帳號名稱" />
        </Form.Item>
        <Form.Item
          name="ver_code"
          rules={[{ required: true, message: '請輸入驗證碼!' }]}
        >
          <Flex gap='small'>
          <Input prefix={<UserOutlined />} placeholder="驗證碼" />
          <Button type="primary" onClick={sendCheckMsg} loading={loading}>
            取得驗證碼
          </Button>
      </Flex>
        </Form.Item>
        <Form.Item
          name="password"
          rules={[{ required: true, message: '請輸入密碼!' }]}
        >
          <Input prefix={<LockOutlined />} type="password" placeholder="密碼" />
        </Form.Item>

        <Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            loading ={loading}
          // className={styles.register_form_button}
          >
            註冊
          </Button>
          <Button className={styles.register_btn_back}>
            <Link to="/login" > 返回</Link>
          </Button>

        </Form.Item>
        <Form.Item>
        </Form.Item>
      </Form>
    </div>
  )
}

export default Register
