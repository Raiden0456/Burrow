import { createBrowserRouter } from "react-router"
import { LoginPage } from "@/features/auth"
import { DashboardPage } from "@/features/dashboard"
import { AppLayout } from "./layout/AppLayout"

export const router = createBrowserRouter([
  { path: "/login", Component: LoginPage },
  {
    Component: AppLayout,
    children: [
      { index: true, Component: DashboardPage },
      { path: "topics", lazy: async () => ({ Component: (await import("@/features/topics")).TopicsPage }) },
      { path: "topics/:slug", lazy: async () => ({ Component: (await import("@/features/topics")).TopicPage }) },
      { path: "quizzes", lazy: async () => ({ Component: (await import("@/features/quiz")).QuizzesPage }) },
      { path: "quiz/:id", lazy: async () => ({ Component: (await import("@/features/quiz")).QuizPage }) },
      { path: "drills/:id", lazy: async () => ({ Component: (await import("@/features/drill/components/DrillPage")).DrillPage }) },
      { path: "chat/:id?", lazy: async () => ({ Component: (await import("@/features/chat")).ChatPage }) },
      { path: "settings", lazy: async () => ({ Component: (await import("@/features/settings")).SettingsPage }) },
    ],
  },
])
