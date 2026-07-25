import GroupDetailScreen from '@/features/groups/group-detail-screen'

export default async function GroupDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  return <GroupDetailScreen params={{ id }} />
}
