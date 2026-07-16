import { notFound } from 'next/navigation'
import AttendanceScreen from '@/features/more/attendance-screen'
import WellbeingScreen from '@/features/more/wellbeing-screen'
import ContactScreen from '@/features/more/contact-screen'
import DebtorsScreen from '@/features/more/debtors-screen'
import WaitingListScreen from '@/features/more/waiting-list-screen'
import SocialScreen from '@/features/more/social-screen'
import type { ExtraScreen } from '@/features/more/more-screen'

const SCREENS: Record<ExtraScreen, React.ComponentType> = {
  attendance: AttendanceScreen,
  wellbeing: WellbeingScreen,
  contact: ContactScreen,
  debtors: DebtorsScreen,
  waiting: WaitingListScreen,
  social: SocialScreen,
}

export default async function MoreScreenPage({ params }: { params: Promise<{ screen: string }> }) {
  const { screen } = await params
  const Screen = SCREENS[screen as ExtraScreen]
  if (!Screen) notFound()
  return <Screen />
}
