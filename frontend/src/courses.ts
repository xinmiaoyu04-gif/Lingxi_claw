// 课程列表（当前为前端静态数据，后端暂无课程接口）。
// 用于 Home / Courses / CourseSpace 之间保持一致的可跳转课程 ID。
export interface Course {
  id: string
  name: string
}

export const courses: Course[] = [
  { id: 'gaoshu', name: '高等数学' },
  { id: 'gailvlun', name: '概率论' },
  { id: 'wuli', name: '大学物理' },
  { id: 'xiandai', name: '线性代数' },
]

export function courseName(id: string): string {
  return courses.find((c) => c.id === id)?.name ?? id
}
