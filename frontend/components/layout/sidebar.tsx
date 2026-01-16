'use client';

import Link from 'next/link';
import { usePathname } from 'next/navigation';
import { motion } from 'framer-motion';
import { Home, Server, Box, Settings, LogOut, Terminal } from 'lucide-react';
import { clsx } from 'clsx';
import { 
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip"

const menuItems = [
  { icon: Home, label: 'Cockpit', href: '/' },
  { icon: Box, label: 'Containers', href: '/containers' },
  { icon: Terminal, label: 'Console', href: '/console' },
  { icon: Settings, label: 'Settings', href: '/settings' },
];

export function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="fixed left-4 top-1/2 -translate-y-1/2 z-50 hidden md:flex flex-col gap-4">
       {/* Glass Container */}
       <motion.div 
         initial={{ x: -100, opacity: 0 }}
         animate={{ x: 0, opacity: 1 }}
         className="glass p-3 rounded-full flex flex-col gap-4 items-center"
       >
          <div className="p-2 bg-gradient-to-br from-indigo-500 to-purple-500 rounded-full shadow-lg shadow-indigo-500/20 mb-2">
            <Server className="w-5 h-5 text-white" />
          </div>

          <TooltipProvider delayDuration={0}>
            {menuItems.map((item) => {
              const isActive = pathname === item.href;
              return (
                <Tooltip key={item.href}>
                  <TooltipTrigger asChild>
                    <Link href={item.href}>
                      <motion.div
                        whileHover={{ scale: 1.1 }}
                        whileTap={{ scale: 0.95 }}
                        className={clsx(
                          "w-10 h-10 rounded-full flex items-center justify-center transition-all duration-300 relative",
                          isActive 
                            ? "bg-white text-black shadow-lg shadow-white/20" 
                            : "text-zinc-400 hover:text-white hover:bg-white/10"
                        )}
                      >
                        <item.icon className="w-5 h-5" />
                        {isActive && (
                           <motion.div 
                             layoutId="sidebar-bubble"
                             className="absolute inset-0 bg-white rounded-full -z-10"
                             transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
                           />
                        )}
                      </motion.div>
                    </Link>
                  </TooltipTrigger>
                  <TooltipContent side="right" className="bg-zinc-900 border-zinc-800 text-zinc-300 ml-4">
                    <p>{item.label}</p>
                  </TooltipContent>
                </Tooltip>
              );
            })}
          </TooltipProvider>

          <div className="h-px w-8 bg-white/10 my-1" />

          <button className="w-10 h-10 rounded-full flex items-center justify-center text-red-400 hover:bg-red-500/10 hover:text-red-300 transition-colors">
            <LogOut className="w-5 h-5" />
          </button>
       </motion.div>
    </aside>
  );
}
