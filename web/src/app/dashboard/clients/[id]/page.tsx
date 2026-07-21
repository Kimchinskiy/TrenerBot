import ClientCardScreen from '@/features/athletes/client-card-screen'

export default async function ClientCardPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  return <ClientCardScreen params={{ id }} />
}
