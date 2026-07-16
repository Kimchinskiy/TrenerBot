import * as React from "react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Home, Menu, MessageSquare, Plus, Settings, Users } from "lucide-react";

export type FloatingNavItem = {
  label: string;
  icon: React.ReactNode;
  href?: string;
  onClick?: () => void;
};

export default function Floatingnavbar({
  items,
}: {
  items: FloatingNavItem[];
}) {
  return (
    <div className="fixed bottom-6 left-0 right-0 flex justify-center z-50">
      <nav className="flex items-center justify-center space-x-4 rounded-full border bg-background p-2 shadow-lg">
        {items.map((item) => (
          <Button
            key={item.label}
            variant="ghost"
            size="icon"
            className="rounded-full"
            onClick={item.onClick}
          >
            {item.icon}
            <span className="sr-only">{item.label}</span>
          </Button>
        ))}
      </nav>
    </div>
  );
}
