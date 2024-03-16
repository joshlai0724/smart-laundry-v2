import { useNavigate, useParams } from 'react-router-dom'
import { Button, Form, Input } from 'antd'
import { LockOutlined } from '@ant-design/icons'
import styles from './index.module.scss'
import { useState } from 'react'
import axios from 'axios'

const ForgotPassword = () => {
  const navigate = useNavigate()
  // const [loading,  setLoading] = useState<boolean>(false)
  const [message, setMessage] = useState<string>('')
  const [form] = Form.useForm()
  const params = useParams()

  interface FieldType {
    password: string
    confirm: string
  }
  console.log(params.verCode);

  const onFinish = async (values: FieldType) => {
    console.log('Success:', values);
    const { password } = values
    // setLoading(true)
    // console.log(password)
    axios.post('http://localhost:80/api/v1/users/.reset-password', {
      new_password: password,
      ver_code: params.verCode
    })
      .then((response) => {
        console.log(response);
        navigate('/')
      }, (error) => {
        const resMessage =
          (error.response?.data?.message) ||
          error.message ||
          error.toString()

        // setLoading(false)
        setMessage(resMessage);
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
        // initialValues={{ remember: true }}
        onFinish={onFinish}
      >
        <Form.Item
          name="password"
          rules={[
            {
              required: true,
              message: '請輸入密碼!',
            },
          ]}
          hasFeedback
        >
          <Input.Password prefix={<LockOutlined />} placeholder="密碼" />
        </Form.Item>
        <Form.Item
          name="confirm"
          dependencies={['password']}
          hasFeedback
          rules={[
            {
              required: true,
              message: '請再輸入一次密碼!',
            },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('password') === value) {
                  return Promise.resolve();
                }
                return Promise.reject(new Error('密碼不一樣!'));
              },
            }),
          ]}
        >
          <Input.Password prefix={<LockOutlined />} placeholder="確認密碼" />
        </Form.Item>
        {message && (
          <div className="form-group">
            <div className="alert alert-danger" role="alert">
              {message}
            </div>
          </div>
        )}
        <Form.Item>
          <Button
            type="primary"
            htmlType="submit"
          // className={styles.register_form_button}
          >
            重設密碼
          </Button>

        </Form.Item>
        <Form.Item>
        </Form.Item>
      </Form>
    </div>
  )
}

export default ForgotPassword
