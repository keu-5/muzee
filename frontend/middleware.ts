import { jwtDecode } from "@/lib/jwt";
import { LINK, PUBLIC_LINKS } from "@/lib/links";
import type { NextRequest } from "next/server";
import { NextResponse } from "next/server";

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;
  const token = request.cookies.get("access_token")?.value;

  const isPublic = PUBLIC_LINKS.some((path) => pathname.startsWith(path));

  // 未ログイン → ログインへ
  if (!token && !isPublic) {
    return NextResponse.redirect(new URL(LINK.login, request.url));
  }

  // ログイン済み → トップへ
  if (token && pathname.startsWith(LINK.login)) {
    return NextResponse.redirect(new URL(LINK.home, request.url));
  }

  // has_profile チェック
  if (token) {
    try {
      const decoded = jwtDecode(token);

      if (!decoded.has_profile && !pathname.startsWith(LINK.createProfile)) {
        return NextResponse.redirect(new URL(LINK.createProfile, request.url));
      }

      if (decoded.has_profile && pathname.startsWith(LINK.createProfile)) {
        return NextResponse.redirect(new URL(LINK.home, request.url));
      }
    } catch (e) {
      console.error("Failed to decode token in middleware:", e);
    }
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next|api|static|favicon.ico).*)"],
};
