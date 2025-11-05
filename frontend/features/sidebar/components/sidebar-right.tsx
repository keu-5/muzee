"use client";

import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Search } from "lucide-react";

const trendingCategories = [
  { id: 1, name: "現代美術", count: 1234 },
  { id: 2, name: "写真", count: 892 },
  { id: 3, name: "デジタルアート", count: 756 },
  { id: 4, name: "映像作品", count: 645 },
  { id: 5, name: "彫刻", count: 534 },
];

const recommendedExhibitions = [
  {
    id: 1,
    title: "都市の光景",
    curator: "Sarah Johnson",
    image: "/urban-photography-street.png",
    likes: 324,
  },
  {
    id: 2,
    title: "デジタル詩集",
    curator: "Alex Chen",
    image: "/digital-art-poetry.jpg",
    likes: 267,
  },
  {
    id: 3,
    title: "音の庭園",
    curator: "Emma Davis",
    image: "/sound-garden-visualization.jpg",
    likes: 198,
  },
];

export const SidebarRight = () => {
  return (
    <aside className="hidden w-80 border-l border-border bg-background/50 h-screen sticky top-0 overflow-y-auto lg:block">
      <div className="p-6 space-y-6">
        {/* Search */}
        <div className="relative">
          <Search className="absolute left-3 top-3 w-4 h-4 text-muted-foreground" />
          <Input
            placeholder="展示やキュレーターを検索"
            className="pl-10 bg-background border-border"
          />
        </div>

        {/* Trending Categories */}
        <div>
          <h3 className="font-semibold text-sm mb-4">トレンドのカテゴリー</h3>
          <div className="space-y-2">
            {trendingCategories.map((category) => (
              <Button
                key={category.id}
                variant="ghost"
                className="w-full justify-between text-left text-sm hover:bg-accent"
              >
                <span>#{category.name}</span>
                <span className="text-xs text-muted-foreground">
                  {category.count}
                </span>
              </Button>
            ))}
          </div>
        </div>

        {/* Recommended Exhibitions */}
        <div>
          <h3 className="font-semibold text-sm mb-4">おすすめの展示</h3>
          <div className="space-y-3">
            {recommendedExhibitions.map((exhibition) => (
              <Card
                key={exhibition.id}
                className="p-3 hover:bg-accent/50 transition-colors cursor-pointer"
              >
                <div className="flex gap-3">
                  <img
                    src={exhibition.image || "/placeholder.svg"}
                    alt={exhibition.title}
                    className="w-16 h-16 rounded object-cover"
                  />
                  <div className="flex-1 min-w-0">
                    <p className="font-medium text-sm line-clamp-2">
                      {exhibition.title}
                    </p>
                    <p className="text-xs text-muted-foreground mb-2">
                      {exhibition.curator}
                    </p>
                    <div className="flex items-center gap-1">
                      <span className="text-xs text-muted-foreground">
                        ❤️ {exhibition.likes}
                      </span>
                    </div>
                  </div>
                </div>
              </Card>
            ))}
          </div>
        </div>

        {/* Footer Info */}
        <div className="border-t border-border pt-6">
          <div className="text-xs text-muted-foreground space-y-2">
            <div>© 2025 Muzee</div>
            <div className="flex gap-2">
              <a href="#" className="hover:text-primary">
                利用規約
              </a>
              <a href="#" className="hover:text-primary">
                プライバシー
              </a>
            </div>
          </div>
        </div>
      </div>
    </aside>
  );
};
