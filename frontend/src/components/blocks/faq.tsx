import Link from "next/link";

import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from "@/components/ui/accordion";
import { cn } from "@/lib/utils";

const categories = [
  {
    title: "Getting Started",
    questions: [
      {
        question: "How do I version my PostgreSQL schema?",
        answer:
          "Connect your PostgreSQL database or push SQL migration files to a repository. Mainline tracks every change as a commit, giving you full Git-like history for your schema. You can branch, diff, review, and roll back any change instantly.",
      },
      {
        question: "Can I review schema changes before deploying?",
        answer:
          "Yes. Every migration goes through a review workflow where your team can see diffs, leave comments, and approve or reject changes before they hit production. This catches breaking changes before they cause downtime.",
      },
      {
        question: "How does schema diffing work?",
        answer:
          "Mainline compares your target schema against the current live state and generates a detailed diff added, removed, or altered tables, columns, indexes, constraints, and more. You can see exactly what will change before applying any migration.",
      },
    ],
  },
  {
    title: "Deployment & CI/CD",
    questions: [
      {
        question: "Can I integrate Mainline with my CI/CD pipeline?",
        answer:
          "Absolutely. Mainline provides a CLI and API that plugs into any CI/CD system. Run schema diffs in pull requests, auto-approve safe migrations, and block dangerous ones. Works with GitHub Actions, GitLab CI, and more.",
      },
      {
        question: "What happens if a migration fails?",
        answer:
          "Mainline wraps each migration in a transaction with automated rollback support. If a migration fails mid-way, we roll back cleanly and alert your team. You can inspect the failure, fix the issue, and retry with confidence.",
      },
    ],
  },
  {
    title: "Security & Compliance",
    questions: [
      {
        question: "Is Mainline SOC 2 compliant?",
        answer:
          "Yes. Mainline is SOC 2 Type II certified. We encrypt data at rest and in transit, provide full audit logs of all schema changes, and support role-based access control (RBAC) to enforce least-privilege access to your production schemas.",
      },
      {
        question: "How do you handle schema drift?",
        answer:
          "Mainline continuously monitors your live databases and compares them against your versioned schema. If drift is detected someone ran an ad-hoc ALTER TABLE, for example we alert your team immediately with a detailed report of what changed.",
      },
    ],
  },
];

export const FAQ = ({
  headerTag = "h2",
  className,
  className2,
}: {
  headerTag?: "h1" | "h2";
  className?: string;
  className2?: string;
}) => {
  return (
    <section className={cn("py-28 lg:py-32", className)}>
      <div className="container max-w-5xl">
        <div className={cn("mx-auto grid gap-16 lg:grid-cols-2", className2)}>
          <div className="space-y-4">
            {headerTag === "h1" ? (
              <h1 className="text-2xl tracking-tight md:text-4xl lg:text-5xl">
                Got Questions?
              </h1>
            ) : (
              <h2 className="text-2xl tracking-tight md:text-4xl lg:text-5xl">
                Got Questions?
              </h2>
            )}
            <p className="text-muted-foreground max-w-md leading-snug lg:mx-auto">
              If you can't find what you're looking for,{" "}
              <Link href="/contact" className="underline underline-offset-4">
                get in touch
              </Link>
              .
            </p>
          </div>

          <div className="grid gap-6 text-start">
            {categories.map((category, categoryIndex) => (
              <div key={category.title} className="">
                <h3 className="text-muted-foreground border-b py-4">
                  {category.title}
                </h3>
                <Accordion type="single" collapsible className="w-full">
                  {category.questions.map((item, i) => (
                    <AccordionItem key={i} value={`${categoryIndex}-${i}`}>
                      <AccordionTrigger>{item.question}</AccordionTrigger>
                      <AccordionContent className="text-muted-foreground">
                        {item.answer}
                      </AccordionContent>
                    </AccordionItem>
                  ))}
                </Accordion>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
};
