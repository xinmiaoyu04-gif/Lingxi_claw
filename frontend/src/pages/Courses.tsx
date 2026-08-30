import PageHeader from '../components/PageHeader'
import Card from '../components/Card'

export default function Courses() {
  return (
    <div>
      <PageHeader title="我的课程" description="管理你的所有课程" />
      <div style={{ display: 'grid', gap: '16px' }}>
        <Card>
          <div>高等数学</div>
        </Card>
        <Card>
          <div>大学物理</div>
        </Card>
        <Card>
          <div>线性代数</div>
        </Card>
      </div>
      <div style={{ marginTop: '16px' }}>
        <button>+ 添加课程</button>
      </div>
    </div>
  )
}
