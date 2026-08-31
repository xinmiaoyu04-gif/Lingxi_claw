import { useNavigate } from 'react-router-dom'
import PageHeader from '../components/PageHeader'
import Card from '../components/Card'

export default function Settings() {
  const navigate = useNavigate()
  return (
    <div>
      <PageHeader title="设置" description="管理你的偏好与账户" />
      <div style={{ display: 'grid', gap: '16px' }}>
        <Card>
          <div>学习画像</div>
        </Card>
        <Card onClick={() => navigate('/agent-settings')}>
          <div>AI 偏好</div>
        </Card>
        <Card>
          <div>通用设置</div>
        </Card>
      </div>
    </div>
  )
}
