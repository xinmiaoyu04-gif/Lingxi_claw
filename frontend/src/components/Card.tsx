type CardProps = {
  children: React.ReactNode
  onClick?: () => void
}

export default function Card({ children, onClick }: CardProps) {
  return (
    <div className="card" onClick={onClick} style={onClick ? { cursor: 'pointer' } : undefined}>
      {children}
    </div>
  )
}
