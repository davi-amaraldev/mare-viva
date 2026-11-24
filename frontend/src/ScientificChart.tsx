import {
  Area,
  AreaChart,
  CartesianGrid,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'

type Point = { name: string; value: number }

export default function ScientificChart({ data }: { data: Point[] }) {
  return (
    <ResponsiveContainer width="100%" height={280}>
      <AreaChart data={data}>
        <CartesianGrid strokeDasharray="3 3" stroke="#29414c" />
        <XAxis dataKey="name" stroke="#91a6ad" />
        <YAxis stroke="#91a6ad" />
        <Tooltip
          contentStyle={{ background: '#102a36', border: '1px solid #35505c', color: '#fff' }}
        />
        <Area dataKey="value" stroke="#14b8a6" fill="#14b8a644" name="Índice relativo" />
      </AreaChart>
    </ResponsiveContainer>
  )
}
