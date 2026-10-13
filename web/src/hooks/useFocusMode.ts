import { matchPath, useLocation } from "react-router"

const focusRoutes = ["/quiz/:id", "/drills/:id", "/chat/:id"]

export function useFocusMode() {
  const { pathname } = useLocation()
  return focusRoutes.some((pattern) => matchPath(pattern, pathname))
}
