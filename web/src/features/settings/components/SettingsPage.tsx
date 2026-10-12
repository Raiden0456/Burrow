import { PageHeader } from "@/components/PageHeader"
import { useI18n } from "@/i18n"
import { InterfaceCard } from "./InterfaceCard"
import { PreferencesCard } from "./PreferencesCard"
import { ProvidersCard } from "./ProvidersCard"
import { PushCard } from "./PushCard"
import { RoutingCard } from "./RoutingCard"

export function SettingsPage() {
  const { t } = useI18n()
  return (
    <>
      <PageHeader title={t("settings.title")} />
      <div className="flex flex-col gap-6">
        <ProvidersCard />
        <RoutingCard />
        <PreferencesCard />
        <InterfaceCard />
        <PushCard />
      </div>
    </>
  )
}
