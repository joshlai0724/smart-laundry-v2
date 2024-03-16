import { Link, useNavigate } from 'react-router-dom'
import { Button, Form, Input } from 'antd'
import { UserOutlined } from '@ant-design/icons'
import styles from './index.module.scss'
import { useState } from 'react'
import axios from 'axios'
import toast from 'react-hot-toast'

const ForgotPassword = () => {
  const navigate = useNavigate()
  const [loading,  setLoading] = useState<boolean>(false)
  const [errMessage, setMessage] = useState<string>('')
  const [form] = Form.useForm()

  interface FieldType {
    phoneNumber: string // 用手機先取得驗證鏈接
  }

  const onFinish = async (values: FieldType) => {
    setLoading(true);
    console.log('Success:', values)
    const phone = form.getFieldValue('phone_number')
    console.log(phone)
    const getMsg = axios.post('http://localhost:80/api/v1/users/' + 'send-reset-password-msg', {
      phone_number: phone
    });
    toast.promise(getMsg,{
      loading: 'Loading',
      success: '重設密碼鏈接已送出',
      // success: '成功登入',
      error: (err) => `發送失敗: ${err.toString()}`
    })
    getMsg.then(
      () => {
        setLoading(false)
        navigate("/login");        
      },
      (err) => {
        const resMessage =
          (err.response?.data?.message) ||
          err.message ||
          err.toString()
        setLoading(false)
        setMessage(resMessage)
        console.log(resMessage);
      }
    )
  }

  return (
    <div className={styles.register_page}>
      <Form
        form={form}
        name="basic"
        className={styles.register_form}
        onFinish={onFinish}
      >
        <Form.Item
          name="phone_number"
          rules={[{ required: true, message: '請輸入手機號碼!' }]}
        >
          <Input prefix={<UserOutlined />} placeholder="手機號碼" />
        </Form.Item>
        <Form.Item>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
          >
            發送重設密碼鏈接
          </Button>
          <Button className={styles.register_btn_back} >
            <Link to="/login" > 返回</Link>
          </Button>
          {errMessage && (
          <div className="form-group">
            <div className="alert alert-danger" role="alert">
              {errMessage}
            </div>
          </div>
        )}
        </Form.Item>
        <Form.Item>
        </Form.Item>
      </Form>
    </div>
  )
}

export default ForgotPassword
