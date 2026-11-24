import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { BrowserRouter, Link, NavLink, Route, Routes, useNavigate } from 'react-router-dom'
import { lazy, Suspense, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import {
  Anchor,
  ArrowDown,
  ArrowRight,
  BarChart3,
  Check,
  ChevronDown,
  Compass,
  Fish,
  Menu,
  Search,
  Shield,
  Waves,
  X,
} from 'lucide-react'
import './style.css'

const ScientificChart = lazy(() => import('./ScientificChart'))

const api = async (path: string, init: RequestInit = {}, retried = false) => {
  const csrf = document.cookie
    .split('; ')
    .find((x) => x.startsWith('mare_csrf='))
    ?.split('=')[1]
  const request = () =>
    fetch(path, {
      ...init,
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
        ...(csrf ? { 'X-CSRF-Token': decodeURIComponent(csrf) } : {}),
        ...init.headers,
      },
    })
  const res = await request()
  if (res.status === 401 && !retried && !path.startsWith('/api/auth/')) {
    const renewed = await fetch('/api/auth/refresh', { method: 'POST', credentials: 'include' })
    if (renewed.ok) return api(path, init, true)
  }
  const body = await res.json()
  if (!res.ok) throw new Error(body.error || 'Falha na solicitação')
  return body
}
type Species = {
  id: number
  common_name: string
  scientific_name: string
  image_url: string
  description: string
  physical_characteristics: string
  size: string
  diet: string
  habitat: string
  distribution: string
  brazil_presence: string
  northeast_presence: string
  ecology: string
  threats: string
  brazil_status: string
  iucn_status: string
  national_assessment_year?: number
  iucn_assessment_year?: number
  iucn_source_url?: string
  national_source_url?: string
  taxonomic_source_url?: string
  source_note?: string
}
function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Home />} />
        <Route path="/especies" element={<Catalog />} />
        <Route path="/observatorio" element={<Observatory />} />
        <Route path="/nordeste" element={<Northeast />} />
        <Route path="/participar" element={<Citizen />} />
        <Route path="/quiz" element={<Quiz />} />
        <Route path="/sobre" element={<About />} />
        <Route path="/admin/login" element={<Login />} />
        <Route path="/admin/*" element={<Admin />} />
      </Routes>
    </BrowserRouter>
  )
}
function Header() {
  const [open, setOpen] = useState(false)
  return (
    <header className="header">
      <Link to="/" className="brand">
        <span className="brand-mark">
          <Waves size={21} />
        </span>
        <span>
          maré<span className="brand-light">viva</span>
          <small>OBSERVATÓRIO DE TUBARÕES</small>
        </span>
      </Link>
      <button className="menu-button" aria-label="Abrir navegação" onClick={() => setOpen(!open)}>
        {open ? <X /> : <Menu />}
      </button>
      <nav className={open ? 'nav open' : 'nav'}>
        {[
          ['Espécies', '/especies'],
          ['Observatório', '/observatorio'],
          ['Nordeste', '/nordeste'],
          ['Quiz', '/quiz'],
          ['Sobre', '/sobre'],
        ].map(([n, p]) => (
          <NavLink key={p} to={p} onClick={() => setOpen(false)}>
            {n}
          </NavLink>
        ))}
        <Link className="nav-cta" to="/participar">
          Participar <ArrowRight size={14} />
        </Link>
      </nav>
    </header>
  )
}
function Footer() {
  return (
    <footer className="footer">
      <Link to="/" className="brand">
        <span className="brand-mark">
          <Waves size={20} />
        </span>
        <span>
          maré<span className="brand-light">viva</span>
        </span>
      </Link>
      <p>Proteger os tubarões é proteger os oceanos.</p>
      <small>Projeto acadêmico independente · Dados científicos com fontes identificadas.</small>
      <div className="footer-links">
        <Link to="/sobre">Sobre o projeto</Link>
        <Link to="/participar">Enviar relato</Link>
        <Link to="/admin/login">Área administrativa</Link>
      </div>
    </footer>
  )
}
function Home() {
  const { data } = useQuery({ queryKey: ['species'], queryFn: () => api('/api/species') })
  const species: Species[] = data?.data || []
  return (
    <>
      <Header />
      <main>
        <section className="hero">
          <img
            src="/images/hero-tiger-shark.jpg"
            alt="Tubarão-tigre fotografado submerso nas Bahamas; imagem de Albert kok, CC BY-SA 3.0"
          />
          <div className="hero-shade" />
          <div className="hero-content">
            <p className="eyebrow">
              <span /> OBSERVATÓRIO CIENTÍFICO · NORDESTE DO BRASIL
            </p>
            <h1>
              O oceano precisa
              <br />
              de seus <em>guardiões.</em>
            </h1>
            <p className="hero-copy">
              Conhecimento é o primeiro passo para conservar. Explore a diversidade dos tubarões e o
              papel essencial que desempenham nos ecossistemas marinhos.
            </p>
            <div className="hero-actions">
              <Link className="button primary" to="/especies">
                Explorar espécies <ArrowRight size={17} />
              </Link>
              <Link className="button text" to="/sobre">
                Conheça o projeto <ArrowDown size={16} />
              </Link>
            </div>
          </div>
          <div className="hero-caption">
            <span>01 / VIDA NA ÁGUA</span>
            <span>FOTOGRAFIA: TUBARÃO-TIGRE · BAHAMAS</span>
          </div>
          <div className="hero-side">CONSERVAÇÃO COMEÇA COM CONHECIMENTO</div>
        </section>
        <section className="intro section-wrap">
          <div className="section-label">01 — POR QUE IMPORTA</div>
          <div className="intro-grid">
            <h2>
              Conhecer para
              <br />
              <em>conservar.</em>
            </h2>
            <div>
              <p className="lead">
                Tubarões não são monstros. São um grupo diverso de peixes que mantém relações
                complexas com a vida marinha — e enfrenta pressões crescentes causadas por
                atividades humanas.
              </p>
              <p>
                Pesca excessiva, captura incidental, perda de habitat e desinformação ameaçam
                populações em todo o mundo. A ciência ajuda a entender cada espécie e orientar ações
                de conservação.
              </p>
              <Link className="inline-link" to="/observatorio">
                Explore os dados do observatório <ArrowRight size={16} />
              </Link>
            </div>
          </div>
        </section>
        <section className="stat-band">
          <div className="stat-intro">
            <span className="eyebrow">PAN TUBARÕES · ICMBIO · 2025</span>
            <h2>
              O retrato da
              <br />
              <em>conservação.</em>
            </h2>
            <p>
              O PAN nacional trata de tubarões e raias (elasmobrânquios). Seu dado não deve ser
              confundido com estatística exclusiva de tubarões.
            </p>
            <Link to="/observatorio" className="inline-link">
              Ver fonte e contexto <ArrowRight size={16} />
            </Link>
          </div>
          <div className="stat-number">
            <strong>
              39<sup>%</sup>
            </strong>
            <span>das espécies brasileiras de elasmobrânquios estão ameaçadas</span>
            <small>ÂMBITO NACIONAL · TUBARÕES E RAIAS · ICMBIO</small>
          </div>
          <div className="stat-art">
            <div className="ring ring1" />
            <div className="ring ring2" />
            <div className="ring ring3" />
            <Fish size={82} strokeWidth={0.7} />
            <span>VIDA SOB A SUPERFÍCIE</span>
          </div>
        </section>
        <section className="section-wrap importance">
          <div className="section-label">02 — EQUILÍBRIO VIVO</div>
          <div className="importance-grid">
            <div>
              <h2>
                Predadores.
                <br />
                <em>Mas não só.</em>
              </h2>
              <p className="lead">
                A diversidade dos tubarões também é diversidade de funções ecológicas.
              </p>
              <p>
                Algumas espécies ocupam posições de predadores de topo; outras são intermediárias ou
                se alimentam por filtração. Seus efeitos variam conforme a espécie e o ecossistema.
              </p>
              <Link className="inline-link" to="/observatorio">
                Explore o simulador educativo <ArrowRight size={16} />
              </Link>
            </div>
            <div className="foodweb">
              <div className="foodnode node1">
                <span>◌</span>
                <b>ECOSSISTEMA</b>
                <small>recifes · manguezais</small>
              </div>
              <div className="foodline" />
              <div className="foodnode node2">
                <Fish />
                <b>TUBARÕES</b>
                <small>papéis diversos</small>
              </div>
              <div className="foodline" />
              <div className="foodnode node3">
                <span>≋</span>
                <b>TEIA ALIMENTAR</b>
                <small>relações interligadas</small>
              </div>
              <div className="food-note">MODELO CONCEITUAL · NÃO É PREVISÃO</div>
            </div>
          </div>
        </section>
        <section className="species-section">
          <div className="section-wrap">
            <div className="section-head">
              <div>
                <div className="section-label">03 — ESPÉCIES BRASILEIRAS</div>
                <h2>
                  Oceano de <em>diversidade.</em>
                </h2>
              </div>
              <Link className="inline-link" to="/especies">
                Ver catálogo completo <ArrowRight size={16} />
              </Link>
            </div>
            <div className="species-row">
              {species.slice(0, 4).map((s, i) => (
                <Link className="species-card" to="/especies" key={s.id}>
                  <div className="species-photo">
                    <img
                      src={s.image_url}
                      alt="Ilustração editorial genérica de tubarão; não representa padrões específicos da espécie"
                      onError={(e) => {
                        e.currentTarget.src = '/images/shark-silhouette.svg'
                      }}
                    />
                    <span>0{i + 1}</span>
                  </div>
                  <h3>{s.common_name}</h3>
                  <i>{s.scientific_name}</i>
                  <small>BRASIL · {s.iucn_status}</small>
                </Link>
              ))}
            </div>
          </div>
        </section>
        <section className="regional-call">
          <div className="regional-image">
            <img src="/images/reef.svg" alt="Ilustração de recife costeiro do Nordeste" />
          </div>
          <div>
            <span className="eyebrow">PESQUISA EM NOSSO LITORAL</span>
            <h2>
              Nordeste,
              <br />
              <em>mar de ciência.</em>
            </h2>
            <p>
              Do monitoramento na costa pernambucana às pesquisas em Fernando de Noronha,
              iniciativas locais ampliam o conhecimento e aproximam a ciência das comunidades.
            </p>
            <Link className="button primary" to="/nordeste">
              Conhecer iniciativas <ArrowRight size={16} />
            </Link>
          </div>
        </section>
        <section className="ods section-wrap">
          <div className="ods-badge">14</div>
          <div>
            <span className="eyebrow">AGENDA 2030 · OBJETIVO PRINCIPAL</span>
            <h2>
              Vida na <em>água.</em>
            </h2>
            <p>
              O Maré Viva contribui com a educação e a conscientização sobre a conservação e o uso
              sustentável dos oceanos, mares e recursos marinhos.
            </p>
            <div className="ods-tags">
              <span>ODS 14 · VIDA NA ÁGUA</span>
              <span>ODS 4 · EDUCAÇÃO</span>
              <span>ODS 9 · INOVAÇÃO</span>
              <span>ODS 12 · CONSUMO RESPONSÁVEL</span>
            </div>
          </div>
        </section>
        <section className="join">
          <span className="eyebrow">UM OCEANO DE OLHARES</span>
          <h2>
            Ciência também se faz
            <br />
            <em>com participação.</em>
          </h2>
          <p>
            Compartilhe observações ambientais. Cada relato será moderado e identificado como
            ciência cidadã não validada cientificamente.
          </p>
          <Link to="/participar" className="button light">
            Enviar um relato <ArrowRight size={16} />
          </Link>
        </section>
      </main>
      <Footer />
    </>
  )
}
function Catalog() {
  const [search, setSearch] = useState('')
  const { data, isLoading } = useQuery({
    queryKey: ['species'],
    queryFn: () => api('/api/species'),
  })
  const all: Species[] = data?.data || []
  const filtered = all.filter((s) =>
    (s.common_name + ' ' + s.scientific_name).toLowerCase().includes(search.toLowerCase()),
  )
  return (
    <>
      <Header />
      <main className="page">
        <div className="page-top">
          <span className="eyebrow">OBSERVATÓRIO · CATÁLOGO</span>
          <h1>
            Habitantes do <em>azul.</em>
          </h1>
          <p>
            Explore espécies registradas no litoral brasileiro. Classificações globais e nacionais
            aparecem separadas; itens não verificados são identificados.
          </p>
          <label className="search-box">
            <Search size={18} />
            <input
              placeholder="Buscar espécie ou nome científico"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
            />
          </label>
        </div>
        {isLoading ? (
          <p>Carregando catálogo…</p>
        ) : (
          <div className="catalog-grid">
            {filtered.map((s) => (
              <article className="catalog-item" key={s.id}>
                <div className="catalog-image">
                  <img
                    src={s.image_url}
                    alt={s.image_url.includes('shark-silhouette') ? `Ilustração genérica de tubarão; não representa especificamente ${s.common_name}` : `Fotografia identificada de ${s.common_name} (${s.scientific_name})`}
                    onError={(e) => {
                      e.currentTarget.src = '/images/shark-silhouette.svg'
                    }}
                  />
                </div>
                <div className="catalog-body">
                  <span className="eyebrow">ESPÉCIE BRASILEIRA</span>
                  <h2>{s.common_name}</h2>
                  <i>{s.scientific_name}</i>
                  <p>{s.description}</p>
                  <div className="species-facts">
                    <span>
                      <small>TAMANHO</small>
                      {s.size || 'Não informado'}
                    </span>
                    <span>
                      <small>HABITAT</small>
                      {s.habitat || 'Não informado'}
                    </span>
                  </div>
                  <div className="species-detail-facts">
                    <p>
                      <b>Características físicas</b>
                      {s.physical_characteristics || 'Não informadas'}
                    </p>
                    <p>
                      <b>Alimentação</b>
                      {s.diet || 'Não informada'}
                    </p>
                    <p>
                      <b>Distribuição</b>
                      {s.distribution || 'Não informada'}
                    </p>
                    <p>
                      <b>Presença no Brasil</b>
                      {s.brazil_presence || 'Não informada'}
                    </p>
                    <p>
                      <b>Presença no Nordeste</b>
                      {s.northeast_presence || 'Não confirmada neste protótipo'}
                    </p>
                    <p>
                      <b>Importância ecológica</b>
                      {s.ecology || 'Não informada'}
                    </p>
                  </div>
                  <div className="statuses">
                    <span>Lista nacional · {s.national_assessment_year || 'ano não informado'}: {s.brazil_status} {s.national_source_url && <a href={s.national_source_url} target="_blank" rel="noreferrer">MMA ↗</a>}</span>
                    <span>IUCN · {s.iucn_assessment_year || 'ano não informado'}: {s.iucn_status} {s.iucn_source_url && <a href={s.iucn_source_url} target="_blank" rel="noreferrer">IUCN ↗</a>}</span>
                  </div>
                  <p className="science-note">{s.source_note}</p>
                  <p className="threats">
                    <b>Ameaças:</b> {s.threats}
                  </p>
                  <a
                    href={s.iucn_source_url || 'https://www.iucnredlist.org/'}
                    target="_blank"
                    rel="noreferrer"
                    className="inline-link"
                  >
                    Consultar ficha IUCN <ArrowRight size={14} />
                  </a>
                  {s.taxonomic_source_url && <a href={s.taxonomic_source_url} target="_blank" rel="noreferrer" className="inline-link">Ficha taxonômica e biológica ↗</a>}
                </div>
              </article>
            ))}
          </div>
        )}
        <p className="science-note">
          Estados de conservação podem mudar conforme novas avaliações. A base nacional do
          ICMBio/MMA e a lista global da IUCN são avaliações distintas.
        </p>
      </main>
      <Footer />
    </>
  )
}
function Observatory() {
  const [q, setQ] = useState(60)
  const { data: ind } = useQuery({
    queryKey: ['indicators'],
    queryFn: () => api('/api/indicators'),
  })
  const sim = useMutation({
    mutationFn: (v: number) =>
      api('/api/ecosystem/simulate', { method: 'POST', body: JSON.stringify({ sharks: v }) }),
  })
  const result = sim.data?.series || [
    { name: 'Tubarões', value: q },
    { name: 'Peixes médios', value: 100 - q * 0.35 },
    { name: 'Peixes pequenos', value: 50 + q * 0.2 },
    { name: 'Saúde do habitat', value: 45 + q * 0.3 },
  ]
  return (
    <>
      <Header />
      <main className="page">
        <div className="page-top">
          <span className="eyebrow">DADOS · CIÊNCIA · CONTEXTO</span>
          <h1>
            O mar em <em>perspectiva.</em>
          </h1>
          <p>
            Indicadores publicados, com fonte, ano, unidade e abrangência. Dados sobre
            elasmobrânquios incluem tubarões e raias quando indicado.
          </p>
        </div>
        <div className="data-highlight">
          <div>
            <span className="eyebrow">BRASIL · ELASMOBRÂNQUIOS</span>
            <strong>39%</strong>
            <p>
              das espécies de tubarões e raias estão ameaçadas, segundo a página institucional do
              PAN Tubarões do ICMBio.
            </p>
            <small>Fonte: ICMBio · PAN Tubarões · 2025 · abrangência nacional</small>
            <a
              href="https://www.gov.br/icmbio/pt-br/assuntos/centros-de-pesquisa/biodiversidade-marinha-do-sudeste-e-sul/o-que-fazemos/pan-tubaroes"
              target="_blank"
              rel="noreferrer"
              className="inline-link"
            >
              Abrir publicação original <ArrowRight size={14} />
            </a>
          </div>
          <BarChart3 size={46} />
        </div>
        <section className="simulator">
          <div className="section-label">SIMULADOR ECOLÓGICO</div>
          <h2>
            Um sistema de <em>relações.</em>
          </h2>
          <p>
            Altere a população relativa de tubarões para observar como um modelo simplificado
            representa possíveis relações em uma teia alimentar.
          </p>
          <label className="range-label">
            População relativa de tubarões <b>{q}%</b>
            <input
              type="range"
              min="0"
              max="100"
              value={q}
              onChange={(e) => {
                const n = Number(e.target.value)
                setQ(n)
                sim.mutate(n)
              }}
            />
          </label>
          <div className="chart-wrap">
            <Suspense fallback={<p>Preparando visualização…</p>}>
              <ScientificChart data={result} />
            </Suspense>
          </div>
          <small className="model-note">
            Modelo conceitual educativo executado pelo backend. Não é previsão científica, não
            representa um ecossistema real e não generaliza todas as espécies como predadores de
            topo.
          </small>
        </section>
        {ind?.data?.length > 0 && (
          <section className="indicators">
            <h2>Indicadores publicados</h2>
            {ind.data.map((v: any) => (
              <article key={v.id}>
                <b>{v.title}</b>
                <strong>
                  {v.value} {v.unit}
                </strong>
                <span>
                  {v.scope} · {v.year}
                </span>
                <small>{v.description}</small>
                <a href={v.url} target="_blank" rel="noreferrer">
                  {v.source || 'Fonte'} ↗
                </a>
              </article>
            ))}
          </section>
        )}
      </main>
      <Footer />
    </>
  )
}
function Northeast() {
  const {data: institutionData}=useQuery({queryKey:['institutions'],queryFn:()=>api('/api/institutions')})
  return (
    <>
      <Header />
      <main className="page">
        <div className="page-top">
          <span className="eyebrow">PESQUISA REGIONAL · ATLÂNTICO SUL</span>
          <h1>
            Nordeste, mar de <em>conhecimento.</em>
          </h1>
          <p>
            Iniciativas, instituições e tecnologias que ajudam a conhecer tubarões em seus ambientes
            naturais. O Maré Viva não possui vínculo institucional com os projetos citados.
          </p>
        </div>
        <div className="regional-feature">
          <img src="/images/reef.svg" alt="Ilustração de ambiente recifal" />
          <div>
            <span className="eyebrow">PERNAMBUCO · UFRPE</span>
            <h2>Do PROTUBA ao ECOTUBA</h2>
            <p>
              A Administração de Fernando de Noronha informa que o ECOTUBA deriva do Projeto de
              Pesquisa e Monitoramento de Tubarões no Estado de Pernambuco (PROTUBA), desenvolvido
              pela UFRPE entre 2004 e 2014 na costa da Região Metropolitana do Recife. Consulte
              fontes atuais para confirmar o escopo vigente.
            </p>
            <a
              className="inline-link"
              href="https://www.noronha.pe.gov.br/em-noronha-projeto-ecotuba-desenvolve-acoes-educativas-e-monitoramento-de-tubaroes/"
              target="_blank"
              rel="noreferrer"
            >
              Fonte: Administração de Fernando de Noronha <ArrowRight size={15} />
            </a>
          </div>
        </div>
        <div className="tech-grid">
          {[
            ['Telemetria', 'Marcas transmitem informações de deslocamento e uso de habitat.'],
            ['Marcação acústica', 'Receptores submersos registram a passagem de animais marcados.'],
            ['Drones', 'Observação remota de áreas costeiras, conforme condições e protocolos.'],
            ['Sensores ambientais', 'Medições ajudam a contextualizar variações ambientais.'],
          ].map((x) => (
            <article key={x[0]}>
              <Compass />
              <h3>{x[0]}</h3>
              <p>{x[1]}</p>
            </article>
          ))}
        </div>
        <section className="source-panel">
          <h2>Instituições, projetos e fontes</h2>
          <div className="institution-list">{(institutionData?.data||[]).map((item:any)=><article key={item.id}><span className="eyebrow">{item.state} · {item.city}</span><h3>{item.name}</h3><p>{item.description}</p><small>{item.focus}</small><a href={item.website} target="_blank" rel="noreferrer">Website oficial ↗</a></article>)}</div>
        </section>
      </main>
      <Footer />
    </>
  )
}
const reportSchema = z.object({
  state: z.string().length(2, 'Informe a sigla com duas letras.'),
  municipality: z.string().min(2, 'Informe o município.'),
  date: z.string().min(10, 'Informe a data.'),
  category: z.enum(['AVISTAMENTO', 'ANIMAL_MORTO', 'CAPTURA_ACIDENTAL', 'POLUICAO', 'DEGRADACAO']),
  species: z.string().optional(),
  description: z
    .string()
    .min(10, 'Descreva ao menos 10 caracteres.')
    .max(2000, 'Limite de 2.000 caracteres.'),
})
type ReportInput = z.infer<typeof reportSchema>
function Citizen() {
  const [done, setDone] = useState(false)
  const [error, setError] = useState('')
  const form = useForm<ReportInput>({
    resolver: zodResolver(reportSchema),
    defaultValues: { state: '', municipality: '', date: '', species: '', description: '' },
  })
  const mutation = useMutation({
    mutationFn: (d: ReportInput) =>
      api('/api/observations', { method: 'POST', body: JSON.stringify(d) }),
    onSuccess: () => setDone(true),
    onError: (e: Error) => setError(e.message),
  })
  const { data: publicReports } = useQuery({queryKey:['public-reports'],queryFn:()=>api('/api/observations')})
  return (
    <>
      <Header />
      <main className="page form-page">
        <span className="eyebrow">CIÊNCIA CIDADÃ · PARTICIPE</span>
        <h1>
          Um olhar atento
          <br />
          pode <em>contribuir.</em>
        </h1>
        <p>
          Relate ocorrências ambientais no litoral. Os registros iniciam como pendentes e só
          aparecem publicamente após moderação. Moderação não significa validação científica.
        </p>
        {(publicReports?.data || []).some((r:any)=>r.demonstration) && <div className="notice demo-notice"><Shield size={18}/><p>Esta lista inclui registros fictícios de demonstração, claramente identificados. Eles não representam avistamentos reais nem dados científicos.</p></div>}
        {!!publicReports?.data?.length && <section className="public-reports"><h2>Relatos publicados</h2><p>Relatos cidadãos aprovados para exibição; aprovação não é validação científica.</p>{publicReports.data.map((r:any)=><article key={r.id}><span>{r.municipality}, {r.state} · {r.date}</span>{r.demonstration&&<b className="demo-tag">DEMONSTRAÇÃO FICTÍCIA</b>}<p>{r.description}</p></article>)}</section>}
        {done ? (
          <div className="success">
            <Check />
            <h2>Relato recebido.</h2>
            <p>
              Obrigado pela participação. O registro seguirá para moderação e ficará identificado
              como relato cidadão não validado cientificamente.
            </p>
          </div>
        ) : (
          <form
            className="citizen-form"
            onSubmit={form.handleSubmit((values) => mutation.mutate(values))}
          >
            <label>
              Tipo de ocorrência
              <select {...form.register('category')} required>
                <option value="">Selecione</option>
                <option value="AVISTAMENTO">Avistamento de tubarão</option>
                <option value="ANIMAL_MORTO">Animal encontrado morto</option>
                <option value="CAPTURA_ACIDENTAL">Captura acidental</option>
                <option value="POLUICAO">Poluição marinha</option>
                <option value="DEGRADACAO">Degradação ambiental</option>
              </select>
            </label>
            <div className="form-row">
              <label>
                Estado (UF)
                <input {...form.register('state')} maxLength={2} placeholder="PE" required />
              </label>
              <label>
                Município
                <input {...form.register('municipality')} required />
              </label>
            </div>
            <div className="form-row">
              <label>
                Data
                <input
                  {...form.register('date')}
                  type="date"
                  max={new Date().toISOString().slice(0, 10)}
                  required
                />
              </label>
              <label>
                Espécie, se conhecida
                <input {...form.register('species')} placeholder="Opcional" />
              </label>
            </div>
            <label>
              Descrição
              <textarea
                {...form.register('description')}
                minLength={10}
                maxLength={2000}
                rows={5}
                placeholder="Descreva o que observou, sem incluir coordenadas precisas."
                required
              />
            </label>
            {Object.values(form.formState.errors).map((e, i) => (
              <p className="error" key={i}>
                {e?.message}
              </p>
            ))}
            <div className="notice">
              <Shield size={18} />
              <p>
                Não persiga, capture, alimente ou se aproxime de animais. Respeite a sinalização e
                as orientações oficiais de segurança nas praias. Esta plataforma acadêmica não
                encaminha automaticamente ocorrências às autoridades. Não informe coordenadas
                precisas de espécies ameaçadas.
              </p>
            </div>
            {error && <p className="error">{error}</p>}
            <button className="button primary" disabled={mutation.isPending}>
              {mutation.isPending ? 'Enviando…' : 'Enviar relato'} <ArrowRight size={16} />
            </button>
          </form>
        )}
      </main>
      <Footer />
    </>
  )
}
function Quiz() {
  const { data } = useQuery({ queryKey: ['quiz'], queryFn: () => api('/api/quiz') })
  const [answers, setAnswers] = useState<Record<string, number>>({})
  const [result, setResult] = useState<any>()
  const evalM = useMutation({
    mutationFn: () =>
      api('/api/quiz/evaluate', { method: 'POST', body: JSON.stringify({ answers }) }),
    onSuccess: setResult,
  })
  const questions = data?.data || []
  return (
    <>
      <Header />
      <main className="page">
        <div className="page-top">
          <span className="eyebrow">MITOS E VERDADES</span>
          <h1>
            Teste o que <em>é fato.</em>
          </h1>
          <p>
            Trinta perguntas sobre diversidade, ecologia, conservação e segurança. As respostas são
            avaliadas pela API.
          </p>
        </div>
        {result ? (
          <section className="quiz-result">
            <span className="eyebrow">RESULTADO</span>
            <strong>
              {result.score}
              <small> / {result.total}</small>
            </strong>
            <p>
              Conhecimento se constrói com curiosidade. Consulte as fontes para continuar
              aprendendo.
            </p>
            <div>
              {result.results.map((x: any) => (
                <p key={x.question_id}>
                  <b>{x.correct ? 'Correto' : 'Revise'}:</b> {x.explanation} {x.source_url && <a href={x.source_url} target="_blank" rel="noreferrer">Fonte ↗</a>}
                </p>
              ))}
            </div>
            <button
              className="button primary"
              onClick={() => {
                setResult(null)
                setAnswers({})
              }}
            >
              Tentar novamente
            </button>
          </section>
        ) : (
          <div className="quiz-list">
            {questions.map((q: any, i: number) => (
              <fieldset key={q.id}>
                <legend>
                  <span>0{i + 1}</span>
                  {q.prompt}
                </legend>
                {q.options.map((o: any) => (
                  <label key={o.id} className={answers[q.id] === o.id ? 'selected' : ''}>
                    <input
                      type="radio"
                      name={String(q.id)}
                      checked={answers[q.id] === o.id}
                      onChange={() => setAnswers({ ...answers, [q.id]: o.id })}
                    />
                    {o.label}
                  </label>
                ))}
              </fieldset>
            ))}
            <button
              className="button primary"
              onClick={() => evalM.mutate()}
              disabled={!questions.length || Object.keys(answers).length !== questions.length}
            >
              Ver resultado <ArrowRight size={16} />
            </button>
          </div>
        )}
        <div className="notice">
          <Shield size={18} />
          <p>
            Incidentes com tubarões podem oferecer risco. Respeite sempre sinalizações e
            recomendações oficiais de segurança nas praias.
          </p>
        </div>
      </main>
      <Footer />
    </>
  )
}
function About() {
  return (
    <>
      <Header />
      <main className="page">
        <div className="page-top">
          <span className="eyebrow">PROJETO ACADÊMICO · ENGENHARIA DE SOFTWARE</span>
          <h1>
            Uma plataforma pela
            <br />
            <em>vida marinha.</em>
          </h1>
        </div>
        <div className="about-copy">
          <h2>Problema</h2>
          <p>
            A desinformação sobre tubarões, a degradação dos ecossistemas marinhos, a pesca
            excessiva e a falta de conhecimento sobre a importância ecológica dessas espécies
            dificultam sua conservação no Brasil.
          </p>
          <h2>Solução</h2>
          <p>
            O Maré Viva propõe uma plataforma científica, educativa e interativa dedicada à
            conservação dos tubarões, com foco especial no litoral nordestino, nas pesquisas
            regionais e na participação da sociedade.
          </p>
          <h2>Agenda 2030</h2>
          <p>
            ODS 14 — Vida na Água é o eixo principal. O projeto também se relaciona aos ODS 4
            (Educação de Qualidade), 9 (Indústria, Inovação e Infraestrutura) e 12 (Consumo e
            Produção Responsáveis).
          </p>
          <h2>Limitações e transparência</h2>
          <p>
            As classificações nacionais precisam ser conferidas na lista oficial mais atualizada do
            MMA/ICMBio; o protótipo sinaliza “não verificado” quando a conferência por espécie não
            foi concluída. O simulador é conceitual, os relatos cidadães não são validações
            científicas e não há encaminhamento automático às autoridades. Nomes de integrantes,
            instituição e professor são configuráveis e não foram presumidos.
          </p>
          <p>
            Inteligência artificial foi usada como apoio à implementação do protótipo; conteúdo
            científico deve ser revisado por especialistas antes de uso decisório.
          </p>
          <h2>Referências</h2>
          <ul>
            <li>
              <a href="https://sdgs.un.org/goals/goal14">ONU — ODS 14</a>
            </li>
            <li>
              <a href="https://www.gov.br/icmbio/pt-br/assuntos/centros-de-pesquisa/biodiversidade-marinha-do-sudeste-e-sul/o-que-fazemos/pan-tubaroes">
                ICMBio — PAN Tubarões
              </a>
            </li>
            <li>
              <a href="https://www.iucnredlist.org/">IUCN Red List</a>
            </li>
            <li>
              <a href="https://www.ufrpe.br/">UFRPE</a>
            </li>
            <li>
              <a href="https://www.noronha.pe.gov.br/em-noronha-projeto-ecotuba-desenvolve-acoes-educativas-e-monitoramento-de-tubaroes/">
                Administração de Fernando de Noronha — ECOTUBA
              </a>
            </li>
          </ul>
        </div>
      </main>
      <Footer />
    </>
  )
}
function Login() {
  const nav = useNavigate()
  const [error, setError] = useState('')
  const m = useMutation({
    mutationFn: (d: any) => api('/api/auth/login', { method: 'POST', body: JSON.stringify(d) }),
    onSuccess: () => nav('/admin/dashboard'),
    onError: (e: Error) => setError(e.message),
  })
  return (
    <main className="login-shell">
      <Link className="brand" to="/">
        <span className="brand-mark">
          <Waves />
        </span>
        <span>
          maréviva<small>ÁREA ADMINISTRATIVA</small>
        </span>
      </Link>
      <form
        onSubmit={(e) => {
          e.preventDefault()
          m.mutate(Object.fromEntries(new FormData(e.currentTarget)))
        }}
      >
        <span className="eyebrow">ACESSO RESTRITO</span>
        <h1>
          Bem-vindo de
          <br />
          <em>volta.</em>
        </h1>
        <label>
          E-mail
          <input type="email" name="email" autoComplete="username" required />
        </label>
        <label>
          Senha
          <input type="password" name="password" autoComplete="current-password" required />
        </label>
        {error && <p className="error">{error}</p>}
        <button className="button primary" disabled={m.isPending}>
          {m.isPending ? 'Entrando…' : 'Entrar'} <ArrowRight size={16} />
        </button>
        <p className="login-help">
          A conta inicial é configurada pelo administrador através das variáveis de ambiente.
        </p>
      </form>
      <Link to="/">← Voltar ao portal</Link>
    </main>
  )
}
type AdminResource = 'indicators' | 'institutions' | 'sources' | 'quiz' | 'users'
const collectionFields: Record<AdminResource, {key:string;label:string;type?:string;required?:boolean;options?:string[]}[]> = {
  indicators:[{key:'title',label:'Título',required:true},{key:'value',label:'Valor',type:'number',required:true},{key:'unit',label:'Unidade',required:true},{key:'year',label:'Ano',type:'number',required:true},{key:'scope',label:'Abrangência geográfica',required:true},{key:'taxonomic_group',label:'Grupo taxonômico'},{key:'description',label:'Descrição',type:'textarea'},{key:'source_id',label:'ID da fonte',type:'number'},{key:'source_url',label:'URL da publicação',type:'url'},{key:'published',label:'Publicado',type:'checkbox'}],
  institutions:[{key:'name',label:'Nome',required:true},{key:'state',label:'Estado (UF)'},{key:'city',label:'Município'},{key:'focus',label:'Área de atuação'},{key:'website',label:'Website oficial',type:'url'},{key:'image_url',label:'Imagem local/URL',type:'url'},{key:'description',label:'Descrição',type:'textarea'}],
  sources:[{key:'title',label:'Título',required:true},{key:'author',label:'Autor ou instituição'},{key:'year',label:'Ano',type:'number',required:true},{key:'url',label:'URL',type:'url',required:true},{key:'doi',label:'DOI'},{key:'kind',label:'Tipo de publicação'},{key:'description',label:'Descrição',type:'textarea'}],
  users:[{key:'name',label:'Nome',required:true},{key:'email',label:'E-mail',type:'email',required:true},{key:'password',label:'Senha inicial/nova',type:'password'},{key:'role',label:'Papel',type:'select',options:['ADMIN','EDITOR','MODERATOR'],required:true},{key:'active',label:'Conta ativa',type:'checkbox'}],
  quiz:[],
}
function AdminCollection({resource,title}: {resource:AdminResource;title:string}) {
  const qc=useQueryClient();const [editing,setEditing]=useState<any>(null);const [correctIndex,setCorrectIndex]=useState(0)
  const query=useQuery({queryKey:['admin-collection',resource],queryFn:()=>api(`/api/admin/${resource}`)})
  const save=useMutation({mutationFn:({id,payload}:{id?:number;payload:any})=>api(id?`/api/admin/${resource}/${id}`:`/api/admin/${resource}`,{method:id?'PUT':'POST',body:JSON.stringify(payload)}),onSuccess:()=>{setEditing(null);qc.invalidateQueries({queryKey:['admin-collection',resource]});qc.invalidateQueries({queryKey:[resource]});qc.invalidateQueries({queryKey:['indicators']});qc.invalidateQueries({queryKey:['institutions']});qc.invalidateQueries({queryKey:['sources']});qc.invalidateQueries({queryKey:['quiz']})}})
  const remove=useMutation({mutationFn:(id:number)=>api(`/api/admin/${resource}/${id}`,{method:'DELETE'}),onSuccess:()=>qc.invalidateQueries({queryKey:['admin-collection',resource]})})
  const rows=query.data?.data||[]
  const fields=collectionFields[resource]
  const submit=(event:React.FormEvent<HTMLFormElement>)=>{event.preventDefault();const form=new FormData(event.currentTarget);const payload:any={};for(const field of fields){if(field.type==='checkbox')payload[field.key]=form.get(field.key)==='on';else{const value=form.get(field.key);payload[field.key]=field.type==='number'?(value===''?(field.key==='source_id'?null:0):Number(value)):value}}
    if(resource==='quiz'){const options=String(form.get('options')||'').split('\n').map(x=>x.trim()).filter(Boolean);payload.options=options.map((label,index)=>({label,correct:index===correctIndex}));payload.published=form.get('published')==='on'}
    if(resource==='users'&&editing?.id&&!payload.password)delete payload.password
    save.mutate({id:editing?.id,payload})
  }
  return <section className="admin-panel"><div className="panel-head"><div><h2>{title}</h2><p>{rows.length} registros</p></div><button className="approve" onClick={()=>{setCorrectIndex(0);setEditing(resource==='quiz'?{published:true}:resource==='users'?{active:true,role:'EDITOR'}:{published:true})}}>Novo registro</button></div>
    {editing&&<form className="editor-form" onSubmit={submit}>
      {resource==='quiz'?<><label>Pergunta<input name="prompt" defaultValue={editing.prompt||''} required/></label><label>Alternativas (uma por linha)<textarea name="options" rows={5} defaultValue={(editing.options||[{label:'',correct:true},{label:'',correct:false}]).map((o:any)=>o.label).join('\n')} required/></label><label>Índice da alternativa correta (começa em zero)<input type="number" min="0" name="correct_index" defaultValue={Math.max(0,(editing.options||[]).findIndex((o:any)=>o.correct))} onChange={e=>setCorrectIndex(Number(e.target.value))}/></label><label>Explicação educativa<textarea name="explanation" rows={3} defaultValue={editing.explanation||''} required/></label><label>Referência URL<input name="source_url" type="url" defaultValue={editing.source_url||''}/></label><label className="check-label"><input type="checkbox" name="published" defaultChecked={editing.published!==false}/> Publicada</label></>:fields.map(f=><label key={f.key}>{f.label}{f.type==='textarea'?<textarea name={f.key} rows={3} defaultValue={editing[f.key]??''} required={f.required}/>:f.type==='checkbox'?<input type="checkbox" name={f.key} defaultChecked={editing[f.key]??(f.key==='active'||f.key==='published')}/>:f.type==='select'?<select name={f.key} defaultValue={editing[f.key]||f.options?.[0]}>{f.options?.map(x=><option key={x}>{x}</option>)}</select>:<input name={f.key} type={f.type||'text'} defaultValue={f.type==='password'?'':editing[f.key]??''} required={f.required} min={f.type==='number'?0:undefined} placeholder={f.type==='password'&&editing.id?'Em branco mantém a senha atual':undefined}/>}</label>)}
      <div className="editor-actions"><button type="button" onClick={()=>setEditing(null)}>Cancelar</button><button className="approve" disabled={save.isPending}>{save.isPending?'Salvando…':'Salvar'}</button></div>{save.error&&<p className="error">{save.error.message}</p>}
    </form>}
    {query.isLoading?<p>Carregando registros…</p>:<div className="simple-record-list">{rows.map((row:any)=><article key={row.id}><div><b>{resource==='quiz'?row.prompt:row.name||row.title||row.email}</b><small>{resource==='quiz'?row.explanation:resource==='indicators'?`${row.value} ${row.unit} · ${row.year} · ${row.scope}`:resource==='institutions'?`${row.city}${row.state?`, ${row.state}`:''} · ${row.focus}`:resource==='users'?`${row.email} · ${row.role} · ${row.active?'ativa':'inativa'}`:`${row.author||''} · ${row.year||''} · ${row.kind||''}`}</small>{resource==='quiz'&&<small>{(row.options||[]).map((o:any,i:number)=>`${o.correct?'✓ ':''}${i+1}. ${o.label}`).join('  ·  ')}</small>}</div><div><button onClick={()=>{setEditing(row);setCorrectIndex(Math.max(0,(row.options||[]).findIndex((o:any)=>o.correct)))}}>Editar</button><button className="reject" onClick={()=>{if(confirm(resource==='users'?'Desativar esta conta?':'Arquivar/remover este registro?'))remove.mutate(row.id)}}>{resource==='users'||resource==='quiz'?'Desativar/arquivar':'Excluir'}</button></div></article>)}</div>}
    {remove.error&&<p className="error">{remove.error.message}</p>}{!query.isLoading&&!rows.length&&<p className="empty">Nenhum registro cadastrado.</p>}
  </section>
}

function Admin() {
  const qc = useQueryClient()
  const nav = useNavigate()
  const [tab, setTab] = useState('dashboard')
  const [editing, setEditing] = useState<any>(null)
  const [obsState, setObsState] = useState('')
  const [obsStatus, setObsStatus] = useState('')
  const { data: userData, isLoading } = useQuery({
    queryKey: ['me'],
    queryFn: () => api('/api/auth/me'),
    retry: false,
  })
  const logout = useMutation({
    mutationFn: () => api('/api/auth/logout', { method: 'POST' }),
    onSuccess: () => {
      qc.clear()
      nav('/admin/login')
    },
  })
  const { data: dash } = useQuery({
    queryKey: ['dashboard'],
    queryFn: () => api('/api/admin/dashboard'),
    enabled: !!userData,
  })
  const { data: obs, refetch } = useQuery({
    queryKey: ['admin-observations', obsState, obsStatus],
    queryFn: () => api(`/api/admin/observations?${new URLSearchParams({state:obsState,status:obsStatus})}`),
    enabled: !!userData && tab === 'observations',
  })
  const { data: sp, refetch: refetchS } = useQuery({
    queryKey: ['admin-species'],
    queryFn: () => api('/api/admin/species'),
    enabled: !!userData && tab === 'species',
  })
  const saveSpecies = useMutation({
    mutationFn: (x: any) =>
      api(x.id ? `/api/admin/species/${x.id}` : '/api/admin/species', {
        method: x.id ? 'PUT' : 'POST',
        body: JSON.stringify(x),
      }),
    onSuccess: () => {
      setEditing(null)
      qc.invalidateQueries({ queryKey: ['admin-species'] })
      qc.invalidateQueries({ queryKey: ['species'] })
    },
  })
  const archiveSpecies = useMutation({
    mutationFn: (id: number) => api(`/api/admin/species/${id}`, { method: 'DELETE' }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-species'] })
      qc.invalidateQueries({ queryKey: ['species'] })
    },
  })
  const moderation = useMutation({
    mutationFn: ({ id, status, reason }: { id: number; status: string; reason: string }) =>
      api(`/api/admin/observations/${id}/status`, {
        method: 'PATCH',
        body: JSON.stringify({ status, rejection_reason: reason }),
      }),
    onSuccess: () => refetch(),
  })
  if (isLoading) return <div className="loading">Abrindo painel…</div>
  if (!userData)
    return (
      <div className="loading">
        Sessão necessária. <Link to="/admin/login">Entrar</Link>
      </div>
    )
  const labels: any = {
    dashboard: 'Visão geral',
    species: 'Espécies',
    indicators: 'Indicadores',
    institutions: 'Instituições',
    observations: 'Ocorrências',
    sources: 'Referências',
    quiz: 'Quiz educativo',
    users: 'Usuários',
  }
  const role=userData.data?.role
  const navItems=[['dashboard','Visão geral'],['species','Espécies'],['indicators','Indicadores'],['institutions','Instituições'],['observations','Ocorrências'],['sources','Referências'],['quiz','Quiz educativo'],['users','Usuários']].filter(([key])=>key==='dashboard'||(key==='observations'?role==='ADMIN'||role==='MODERATOR':key==='users'?role==='ADMIN':role==='ADMIN'||role==='EDITOR'))
  return (
    <div className="admin-shell">
      <aside className="admin-side">
        <Link to="/" className="brand">
          <span className="brand-mark">
            <Waves />
          </span>
          <span>
            maré<span className="brand-light">viva</span>
            <small>PAINEL CIENTÍFICO</small>
          </span>
        </Link>
        <small className="side-caption">ESPAÇO DE TRABALHO</small>
        {navItems.map(([k, v]) => (
          <button key={k} onClick={() => setTab(k)} className={tab === k ? 'active' : ''}>
            {k === 'dashboard'||k==='indicators'? <BarChart3 /> : k === 'species'||k==='quiz' ? <Fish /> : k==='users'?<Shield/>:<Anchor />}
            {v}
          </button>
        ))}
        <div className="side-user">
          <span>{userData.data?.name?.slice(0, 1) || 'M'}</span>
          <div>
            <b>{userData.data?.name}</b>
            <small>{userData.data?.role}</small>
          </div>
          <button aria-label="Sair" onClick={() => logout.mutate()}>
            ↗
          </button>
        </div>
      </aside>
      <section className="admin-main">
        <header>
          <div>
            <span>Maré Viva</span>
            <ChevronDown size={14} />
            <b>{labels[tab]}</b>
          </div>
          <button onClick={() => logout.mutate()}>Encerrar sessão</button>
        </header>
        <main>
          <div className="admin-title">
            <div>
              <span className="eyebrow">PAINEL DE CONSERVAÇÃO</span>
              <h1>{labels[tab]}</h1>
            </div>
            <div className="admin-user">
              <span>{userData.data?.name?.slice(0, 1)}</span>
              {userData.data?.name}
            </div>
          </div>
          {tab === 'dashboard' ? (
            <>
              <div className="metrics">
                {Object.entries(dash?.data?.counts || {}).map(([k, v]) => (
                  <article key={k}>
                    <small>
                      {
                        (
                          {
                            species: 'ESPÉCIES',
                            threatened_species: 'COM RISCO GLOBAL',
                            observations: 'RELATOS (TOTAL)',
                            community_reports: 'RELATOS DA COMUNIDADE',
                            demonstration_observations: 'DEMONSTRAÇÕES FICTÍCIAS',
                            pending: 'PENDENTES',
                            approved: 'APROVADOS',
                            institutions: 'INSTITUIÇÕES',
                            sources: 'REFERÊNCIAS',
                          } as any
                        )[k]
                      }
                    </small>
                    <strong>{v as number}</strong>
                  </article>
                ))}
              </div>
              <section className="admin-panel">
                <h2>Ocorrências por estado <small>(inclui demonstrações fictícias)</small></h2>
                {(dash?.data?.observations_by_state || []).length ? (
                  dash.data.observations_by_state.map((x: any) => (
                    <div className="barline" key={x.state}>
                      <span>{x.state}</span>
                      <div>
                        <i style={{ width: `${Math.max(8, x.count * 12)}%` }} />
                      </div>
                      <b title={`${x.community_count} relatos da comunidade · ${x.demonstration_count} demonstrações fictícias`}>{x.count}<small> ({x.community_count} + {x.demonstration_count} demo)</small></b>
                    </div>
                  ))
                ) : (
                  <p className="empty">Ainda não há relatos registrados.</p>
                )}
              </section>
            </>
          ) : null}
          {tab === 'observations' && (
            <section className="admin-panel">
              <div className="panel-head">
                <h2>Relatos para moderação</h2>
                <button onClick={() => refetch()}>Atualizar</button>
              </div>
              <div className="observation-filters">
                <label>Estado<select value={obsState} onChange={e=>setObsState(e.target.value)}><option value="">Todos</option>{['PE','PB','RN','CE','BA','AL','SE','MA','PI'].map(x=><option key={x}>{x}</option>)}</select></label>
                <label>Status<select value={obsStatus} onChange={e=>setObsStatus(e.target.value)}><option value="">Todos</option>{['PENDING','UNDER_REVIEW','APPROVED','REJECTED'].map(x=><option key={x}>{x}</option>)}</select></label>
              </div>
              <div className="table-scroll">
                <table>
                  <thead>
                    <tr>
                      <th>LOCAL</th>
                      <th>TIPO</th>
                      <th>DATA</th>
                      <th>STATUS</th>
                      <th>AÇÕES</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(obs?.data || []).map((o: any) => (
                      <tr key={o.id}>
                        <td>
                          <b>
                            {o.municipality}, {o.state}
                          </b>
                          {o.demonstration && <b className="demo-tag">DEMONSTRAÇÃO FICTÍCIA</b>}
                          <small>{o.description}</small>
                        </td>
                        <td>{o.category}</td>
                        <td>{o.date}</td>
                        <td>
                          <span className={`pill ${o.status.toLowerCase()}`}>{o.status}</span>
                        </td>
                        <td>
                          {o.status !== 'APPROVED' && (
                            <button
                              className="approve"
                              onClick={() =>
                                moderation.mutate({ id: o.id, status: 'APPROVED', reason: '' })
                              }
                            >
                              Aprovar
                            </button>
                          )}{' '}
                          {o.status !== 'REJECTED' && (
                            <button
                              className="reject"
                              onClick={() => {
                                const reason = prompt('Motivo da rejeição:')
                                if (reason)
                                  moderation.mutate({ id: o.id, status: 'REJECTED', reason })
                              }}
                            >
                              Rejeitar
                            </button>
                          )}{' '}
                          {o.status === 'PENDING' && (
                            <button
                              onClick={() =>
                                moderation.mutate({ id: o.id, status: 'UNDER_REVIEW', reason: '' })
                              }
                            >
                              Revisar
                            </button>
                          )}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
                {!obs?.data?.length && <p className="empty">Nenhum relato pendente.</p>}
              </div>
              <small className="table-note">
                A aprovação é uma decisão de moderação e não equivale a validação científica.
              </small>
            </section>
          )}
          {tab === 'species' && (
            <section className="admin-panel">
              <div className="panel-head">
                <h2>Catálogo de espécies</h2>
                <div>
                  <button onClick={() => refetchS()}>Atualizar</button>{' '}
                  <button className="approve" onClick={() => setEditing({ published: true })}>
                    Cadastrar espécie
                  </button>
                </div>
              </div>
              {editing !== null && (
                <form
                  className="editor-form"
                  onSubmit={(e) => {
                    e.preventDefault()
                    const f = new FormData(e.currentTarget)
                    saveSpecies.mutate({
                      id: editing.id,
                      common_name: f.get('common_name'),
                      scientific_name: f.get('scientific_name'),
                      image_url: f.get('image_url'),
                      description: f.get('description'),
                      physical_characteristics: f.get('physical_characteristics'),
                      size: f.get('size'),
                      diet: f.get('diet'),
                      habitat: f.get('habitat'),
                      distribution: f.get('distribution'),
                      brazil_presence: f.get('brazil_presence'),
                      northeast_presence: f.get('northeast_presence'),
                      ecology: f.get('ecology'),
                      threats: f.get('threats'),
                      brazil_status: f.get('brazil_status'),
                      iucn_status: f.get('iucn_status'),
                      published: f.get('published') === 'on',
                    })
                  }}
                >
                  <div className="form-row">
                    <label>
                      Nome popular
                      <input name="common_name" defaultValue={editing.common_name || ''} required />
                    </label>
                    <label>
                      Nome científico
                      <input
                        name="scientific_name"
                        defaultValue={editing.scientific_name || ''}
                        required
                      />
                    </label>
                  </div>
                  <div className="form-row">
                    <label>
                      Imagem local ou URL
                      <input
                        name="image_url"
                        defaultValue={editing.image_url || '/images/shark-silhouette.svg'}
                      />
                    </label>
                    <label>
                      Tamanho
                      <input name="size" defaultValue={editing.size || ''} />
                    </label>
                  </div>
                  <div className="form-row">
                    <label>
                      Alimentação
                      <input name="diet" defaultValue={editing.diet || ''} />
                    </label>
                    <label>
                      Habitat
                      <input name="habitat" defaultValue={editing.habitat || ''} />
                    </label>
                  </div>
                  <div className="form-row">
                    <label>
                      Distribuição
                      <input name="distribution" defaultValue={editing.distribution || ''} />
                    </label>
                    <label>
                      Estado nacional (ICMBio)
                      <input
                        name="brazil_status"
                        defaultValue={editing.brazil_status || 'Não verificado'}
                      />
                    </label>
                  </div>
                  <label>
                    Descrição
                    <textarea
                      name="description"
                      rows={3}
                      defaultValue={editing.description || ''}
                    />
                  </label>
                  <div className="form-row">
                    <label>
                      Características físicas
                      <input
                        name="physical_characteristics"
                        defaultValue={editing.physical_characteristics || ''}
                      />
                    </label>
                    <label>
                      Presença no Brasil
                      <input
                        name="brazil_presence"
                        defaultValue={editing.brazil_presence || 'Ocorrência a confirmar'}
                      />
                    </label>
                  </div>
                  <label>
                    Presença no Nordeste
                    <input
                      name="northeast_presence"
                      defaultValue={editing.northeast_presence || 'Ocorrência a confirmar'}
                    />
                  </label>
                  <div className="form-row">
                    <label>
                      Importância ecológica
                      <input name="ecology" defaultValue={editing.ecology || ''} />
                    </label>
                    <label>
                      Ameaças
                      <input name="threats" defaultValue={editing.threats || ''} />
                    </label>
                  </div>
                  <label>
                    Estado global (IUCN)
                    <input
                      name="iucn_status"
                      defaultValue={editing.iucn_status || 'Não verificado'}
                    />
                  </label>
                  <label className="check-label">
                    <input
                      type="checkbox"
                      name="published"
                      defaultChecked={editing.published !== false}
                    />{' '}
                    Publicar no catálogo público
                  </label>
                  <div className="editor-actions">
                    <button type="button" onClick={() => setEditing(null)}>
                      Cancelar
                    </button>
                    <button className="approve" disabled={saveSpecies.isPending}>
                      {saveSpecies.isPending ? 'Salvando…' : 'Salvar espécie'}
                    </button>
                  </div>
                  {saveSpecies.error && <p className="error">{saveSpecies.error.message}</p>}
                </form>
              )}
              <div className="table-scroll">
                <table>
                  <thead>
                    <tr>
                      <th>ESPÉCIE</th>
                      <th>NOME CIENTÍFICO</th>
                      <th>ICMBIO</th>
                      <th>IUCN</th>
                      <th>AÇÕES</th>
                    </tr>
                  </thead>
                  <tbody>
                    {(sp?.data || []).map((s: Species) => (
                      <tr key={s.id}>
                        <td>
                          <b>{s.common_name}</b>
                        </td>
                        <td>
                          <i>{s.scientific_name}</i>
                        </td>
                        <td>{s.brazil_status}</td>
                        <td>{s.iucn_status}</td>
                        <td>
                          <button className="approve" onClick={() => setEditing(s)}>
                            Editar
                          </button>{' '}
                          <button
                            className="reject"
                            onClick={() => {
                              if (confirm(`Arquivar ${s.common_name}?`)) archiveSpecies.mutate(s.id)
                            }}
                          >
                            Arquivar
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              <p className="table-note">
                Status nacional e internacional são campos distintos. Confirme as fontes por espécie
                antes de publicar alterações. Arquivar remove a espécie do catálogo público sem
                apagar o histórico.
              </p>
            </section>
          )}
          {tab === 'indicators'&&<AdminCollection resource="indicators" title="Indicadores científicos"/>}
          {tab === 'institutions'&&<AdminCollection resource="institutions" title="Instituições e projetos"/>}
          {tab === 'sources'&&<AdminCollection resource="sources" title="Referências científicas"/>}
          {tab === 'quiz'&&<AdminCollection resource="quiz" title="Perguntas do quiz"/>}
          {tab === 'users'&&userData.data?.role==='ADMIN'&&<AdminCollection resource="users" title="Usuários administrativos"/>}
        </main>
      </section>
    </div>
  )
}
export default function Root() {
  return <App />
}
