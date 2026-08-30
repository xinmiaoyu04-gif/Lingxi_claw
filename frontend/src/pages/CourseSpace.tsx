import { Link } from 'react-router-dom'
import PageHeader from '../components/PageHeader'
import Card from '../components/Card'

export default function CourseSpace() {
  return (
    <div>
      <Link to="/courses">← 我的课程</Link>
      <PageHeader title="高等数学" description="课程空间" />
      <div style={{ display: 'grid', gap: '16px' }}>
        <Card>
          <div>概览</div>
        </Card>
        <Card>
          <div>课程资料知识库</div>
        </Card>
        <Card>
          <div>AI 作业辅导</div>
        </Card>
        <Card>
          <div>错题记录</div>
        </Card>
        <Card>
          <div>学习分析</div>
        </Card>
      </div>
    </div>
  )
}
