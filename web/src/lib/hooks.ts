import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { endpoints } from './api'
import type { SocialLink } from './types'

export function useMe() {
  return useQuery({ queryKey: ['me'], queryFn: () => endpoints.me() })
}

export function useClients() {
  return useQuery({ queryKey: ['clients'], queryFn: () => endpoints.clients() })
}

export function useSchedule(from: string, to: string) {
  return useQuery({
    queryKey: ['lessons', from, to],
    queryFn: () => endpoints.lessons(from, to),
    enabled: !!from && !!to,
  })
}

export function useScheduleWeek(from: string, to: string) {
  return useQuery({
    queryKey: ['schedule', from, to],
    queryFn: () => endpoints.schedule(from, to),
    enabled: !!from && !!to,
  })
}

export function useAttendance(lessonId: number) {
  return useQuery({
    queryKey: ['attendance', lessonId],
    queryFn: () => endpoints.attendance(lessonId),
    enabled: !!lessonId,
  })
}

export function useMarkAttendance(lessonId: number) {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (vars: { clientId: number; present: boolean }) =>
      endpoints.markAttendance(lessonId, vars.clientId, vars.present),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['attendance', lessonId] }),
  })
}

export function useDebtors(days: number) {
  return useQuery({
    queryKey: ['debtors', days],
    queryFn: () => endpoints.debtors(days),
    enabled: !!days,
  })
}

export function useWaitingList() {
  return useQuery({ queryKey: ['waiting'], queryFn: () => endpoints.waitingList() })
}

export function useAddWaitingList() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (clientId: number) => endpoints.addWaitingList(clientId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['waiting'] }),
  })
}

export function useRemoveWaitingList() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (id: number) => endpoints.removeWaitingList(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['waiting'] }),
  })
}

export function useSubmitWellbeing() {
  return useMutation({
    mutationFn: (vars: { lessonId: number; wellbeing: number; note: string }) =>
      endpoints.wellbeing(vars.lessonId, vars.wellbeing, vars.note),
  })
}

export function useWellbeingHistory(clientId: number) {
  return useQuery({
    queryKey: ['wellbeing', clientId],
    queryFn: () => endpoints.wellbeingHistory(clientId),
    enabled: !!clientId,
  })
}

export function useMessageCoach() {
  return useMutation({
    mutationFn: (vars: { from: string; text: string }) => endpoints.messageCoach(vars.from, vars.text),
  })
}

export function useCreateScheduleEntry() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (data: { client_id: number; date: string; time: string; duration?: number }) =>
      endpoints.createScheduleEntry(data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['schedule'] })
    },
  })
}

export function useSocialMedia() {
  return useQuery({ queryKey: ['social'], queryFn: () => endpoints.socialMedia() })
}

export function useSaveSocialLinks() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (links: SocialLink[]) => endpoints.saveSocialLinks(links),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['social'] }),
  })
}

export function useFaq(q: string) {
  return useQuery({
    queryKey: ['faq', q],
    queryFn: () => endpoints.faq(q),
    enabled: q.trim().length > 0,
  })
}

export function useDateAttendance(date: string) {
  return useQuery({
    queryKey: ['dateAttendance', date],
    queryFn: () => endpoints.dateAttendance(date),
    enabled: !!date,
  })
}

export function useSaveDateAttendance() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (vars: { date: string; entries: { client_id: number; present: boolean }[] }) =>
      endpoints.saveDateAttendance(vars.date, vars.entries),
    onSuccess: (_, vars) => qc.invalidateQueries({ queryKey: ['dateAttendance', vars.date] }),
  })
}

// --- Coach hooks ---
export function useCoachOnboarding() {
  return useQuery({ queryKey: ['coach-onboarding'], queryFn: () => endpoints.coachOnboarding() })
}

export function useUpgradeToCoach() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (vars: { fullName: string; sport: string }) => endpoints.upgradeToCoach(vars.fullName, vars.sport),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['me'] }); qc.invalidateQueries({ queryKey: ['coach-onboarding'] }) },
  })
}

export function useCoachSubscription() {
  return useQuery({ queryKey: ['coach-subscription'], queryFn: () => endpoints.coachSubscription() })
}

export function useStartCoachTrial() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => endpoints.startCoachTrial(),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['coach-subscription'] }),
  })
}

export function useActivateCoachSubscription() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (days: number) => endpoints.activateCoachSubscription(days),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['coach-subscription'] }),
  })
}

// --- Parent hooks ---
export function useUpgradeToParent() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: () => endpoints.upgradeToParent(),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['me'] }) },
  })
}

export function useLinkChild() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (vars: { fullName: string; birthDate: string }) => endpoints.linkChild(vars.fullName, vars.birthDate),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['me'] }) },
  })
}

export function useChildrenLessonStatus() {
  return useQuery({ queryKey: ['children-status'], queryFn: () => endpoints.childrenLessonStatus() })
}

export function useParentNotifPrefs() {
  return useQuery({ queryKey: ['parent-notif-prefs'], queryFn: () => endpoints.parentNotifPrefs() })
}

export function useSaveParentNotifPref() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (vars: { child_id: number; lesson_start?: boolean; lesson_end_15?: boolean; lesson_missed?: boolean }) =>
      endpoints.saveParentNotifPref(vars),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['parent-notif-prefs'] }),
  })
}
