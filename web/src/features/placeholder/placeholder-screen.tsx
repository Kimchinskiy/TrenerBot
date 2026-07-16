import { ScreenHeader, Empty } from '@/components/ui/screen'

export default function Placeholder({ title }: { title: string }) {
  return (
    <div>
      <ScreenHeader title={title} />
      <Empty text="Раздел в разработке" />
    </div>
  )
}
