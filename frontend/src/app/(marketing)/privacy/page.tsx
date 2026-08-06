import { Background } from "@/components/background";
import Privacy from "./privacy.mdx";

const Page = () => {
  return (
    <Background>
      <section className="py-28 lg:pt-44 lg:pb-32">
        <div className="container max-w-3xl">
          <article className="prose prose-lg dark:prose-invert mx-auto">
            <Privacy />
          </article>
        </div>
      </section>
    </Background>
  );
};

export default Page;
