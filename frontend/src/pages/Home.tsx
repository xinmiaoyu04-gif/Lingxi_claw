import PageHeader from '../components/PageHeader'
import Card from '../components/Card'

export default function Home() {
  return (
    <div className="home-page">
      <PageHeader title="学习首页" description="集中查看你的学习进度与最近任务。" />

      <div className="home-top-row">
        <div className="home-section">
          <h2 className="home-section-title">今日学习</h2>
          <Card>
            <div className="home-card-title">今日学习</div>
            <div className="home-card-value">2 / 4 项任务</div>
            <div className="home-card-subtitle">继续保持学习节奏</div>
          </Card>
        </div>

        <div className="home-section">
          <h2 className="home-section-title">最近课程</h2>
          <div className="home-cards">
            <Card>
              <div className="home-card-title">高等数学</div>
            </Card>
            <Card>
              <div className="home-card-title">概率论</div>
            </Card>
            <Card>
              <div className="home-card-title">大学物理</div>
            </Card>
          </div>
        </div>
      </div>

      <div className="home-section">
        <h2 className="home-section-title">快速入口</h2>
        <div className="home-cards home-cards--three">
          <Card>
            <div className="home-card-title">期末突击</div>
          </Card>
          <Card>
            <div className="home-card-title">日常作业</div>
          </Card>
          <Card>
            <div className="home-card-title">其他问题</div>
          </Card>
        </div>
      </div>
    </div>
  )
}
