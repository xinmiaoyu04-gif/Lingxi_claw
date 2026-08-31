import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import PageHeader from '../components/PageHeader'
import Card from '../components/Card'
import { courses } from '../courses'

export default function Courses() {
  const navigate = useNavigate()
  const [notice, setNotice] = useState(false)
  return (
    <div>
      <PageHeader title="我的课程" description="管理你的所有课程" />
      <div style={{ display: 'grid', gap: '16px' }}>
        {courses.map((c) => (
          <Card key={c.id} onClick={() => navigate(`/courses/${c.id}`)}>
            <div>{c.name}</div>
          </Card>
        ))}
      </div>
      <div style={{ marginTop: '16px' }}>
        <button onClick={() => setNotice(true)}>+ 添加课程</button>
        {notice && (
          <div style={{ marginTop: '8px', color: '#5E5D59', fontSize: '13px' }}>
            添加课程需要后端接口，暂未开放。
          </div>
        )}
      </div>
    </div>
  )
}
