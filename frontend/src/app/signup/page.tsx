import Image from "next/image";
import Link from "next/link";

import { FcGoogle } from "react-icons/fc";
import { FaGithub, FaSlack } from "react-icons/fa";

import { Background } from "@/components/background";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Input } from "@/components/ui/input";

const Signup = () => {
  return (
    <Background>
      <section className="py-28 lg:pt-44 lg:pb-32">
        <div className="container">
          <div className="flex flex-col gap-4">
            <Card className="mx-auto w-full max-w-sm">
              <CardHeader className="flex flex-col items-center space-y-0">
                <Image
                  src="/logo.svg"
                  alt="logo"
                  width={94}
                  height={18}
                  className="mb-7 dark:invert"
                />
                <p className="mb-2 text-2xl font-bold">Start your free trial</p>
                <p className="text-muted-foreground">
                  Sign up in less than 2 minutes.
                </p>
              </CardHeader>
              <CardContent>
                <div className="grid gap-4">
                  <Input type="text" placeholder="Enter your name" required />
                  <Input type="email" placeholder="Enter your email" required />
                  <div>
                    <Input
                      type="password"
                      placeholder="Enter your password"
                      required
                    />
                    <p className="text-muted-foreground mt-1 text-sm">
                      Must be at least 8 characters.
                    </p>
                  </div>
                  <Button type="submit" className="mt-2 w-full">
                    Create an account
                  </Button>
                  <div className="relative my-2">
                    <div className="absolute inset-0 flex items-center">
                      <span className="w-full border-t" />
                    </div>
                    <div className="relative flex justify-center text-xs uppercase">
                      <span className="bg-card px-2 text-muted-foreground">
                        Or continue with
                      </span>
                    </div>
                  </div>
                  <div className="grid grid-cols-3 gap-2">
                    <Button variant="outline" className="w-full">
                      <FcGoogle className="size-5" />
                    </Button>
                    <Button variant="outline" className="w-full">
                      <FaGithub className="size-5" />
                    </Button>
                    <Button variant="outline" className="w-full">
                      <FaSlack className="size-5 text-[#4A154B]" />
                    </Button>
                  </div>
                </div>
                <div className="text-muted-foreground mx-auto mt-8 flex justify-center gap-1 text-sm">
                  <p>Already have an account?</p>
                  <Link href="/login" className="text-primary font-medium">
                    Log in
                  </Link>
                </div>
              </CardContent>
            </Card>
          </div>
        </div>
      </section>
    </Background>
  );
};

export default Signup;
