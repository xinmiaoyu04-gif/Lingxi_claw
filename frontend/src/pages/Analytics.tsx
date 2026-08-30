import PageHeader from '../components/PageHeader'
import Card from '../components/Card'

export default function Analytics() {
  return (
    <div>
      <PageHeader title="学习分析" description="查看你的学习数据与趋势" />
      <div style={{ display: 'grid', gap: '16px' }}>
        <Card>
          <div>整体学习情况</div>
        </Card>
        <Card>
          <div>各课程掌握度</div>
        </Card>
        <Card>
          <div>知识薄弱点</div>
        </Card>
        <Card>
          <div>作业表现</div>
        </Card>
        <Card>
          <div>学习趋势</div>
        </Card>
      </div>
    </div>
  )
}
