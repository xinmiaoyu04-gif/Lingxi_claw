import { Link, useParams, useNavigate } from 'react-router-dom'
import PageHeader from '../components/PageHeader'
import Card from '../components/Card'
import { courseName } from '../courses'

export default function CourseSpace() {
  const { courseId = '' } = useParams()
  const navigate = useNavigate()
  const name = courseName(courseId)
  return (
    <div>
      <Link to="/courses">← 我的课程</Link>
      <PageHeader title={name} description="课程空间" />
      <div style={{ display: 'grid', gap: '16px' }}>
        <Card>
          <div>概览</div>
        </Card>
        <Card>
          <div>课程资料知识库</div>
        </Card>
        <Card onClick={() => navigate('/homework')}>
          <div>AI 作业辅导</div>
        </Card>
        <Card>
          <div>错题记录</div>
        </Card>
        <Card onClick={() => navigate('/analytics')}>
          <div>学习分析</div>
        </Card>
      </div>
    </div>
  )
}
